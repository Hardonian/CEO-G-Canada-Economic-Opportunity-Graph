"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import type { ReactNode } from "react";
import { Language, localeFor, translateText } from "@/lib/i18n";

interface LanguageContextValue {
  language: Language;
  locale: "en-CA" | "fr-CA";
  setLanguage: (language: Language) => void;
  toggleLanguage: () => void;
  t: (value: string) => string;
}

const LanguageContext = createContext<LanguageContextValue | null>(null);

function initialLanguage(): Language {
  if (typeof document === "undefined") return "en";
  return document.documentElement.dataset.language === "fr" ? "fr" : "en";
}

export function LanguageProvider({ children }: { children: ReactNode }) {
  // Keep hydration aligned with server-rendered English, then apply the saved
  // preference immediately after mount.
  const [language, updateLanguage] = useState<Language>("en");

  const setLanguage = useCallback((nextLanguage: Language) => {
    updateLanguage(nextLanguage);
    document.documentElement.lang = localeFor(nextLanguage);
    document.documentElement.dataset.language = nextLanguage;
    try {
      localStorage.setItem("cog-language", nextLanguage);
      document.cookie = `cog-language=${nextLanguage}; Path=/; Max-Age=31536000; SameSite=Lax`;
    } catch {
      // Storage can be unavailable in privacy-restricted contexts; the in-page
      // language state and document semantics still update.
    }
  }, []);

  useEffect(() => {
    setLanguage(initialLanguage());
  }, [setLanguage]);

  const value = useMemo<LanguageContextValue>(() => ({
    language,
    locale: localeFor(language),
    setLanguage,
    toggleLanguage: () => setLanguage(language === "en" ? "fr" : "en"),
    t: (text) => translateText(text, language),
  }), [language, setLanguage]);

  return <LanguageContext.Provider value={value}>{children}</LanguageContext.Provider>;
}

export function useLanguage(): LanguageContextValue {
  const value = useContext(LanguageContext);
  if (!value) throw new Error("useLanguage must be used inside LanguageProvider");
  return value;
}

const TRANSLATED_ATTRIBUTES = ["aria-label", "placeholder", "title", "alt"] as const;

function shouldSkip(element: Element | null): boolean {
  return Boolean(element?.closest("[data-no-translate], code, pre, kbd, samp, script, style"));
}

export function LocalizedContent({ children }: { children: ReactNode }) {
  const { language } = useLanguage();
  const rootRef = useRef<HTMLDivElement>(null);
  const textSources = useRef(new WeakMap<Text, string>());
  const attributeSources = useRef(new WeakMap<Element, Map<string, string>>());

  useEffect(() => {
    const root = rootRef.current;
    if (!root) return;
    let applying = false;

    const translateNode = (node: Node) => {
      if (node.nodeType === Node.TEXT_NODE) {
        const textNode = node as Text;
        if (shouldSkip(textNode.parentElement)) return;
        if (!textSources.current.has(textNode)) textSources.current.set(textNode, textNode.nodeValue ?? "");
        const source = textSources.current.get(textNode) ?? "";
        const translated = translateText(source, language);
        if (textNode.nodeValue !== translated) textNode.nodeValue = translated;
        return;
      }
      if (!(node instanceof Element) || shouldSkip(node)) return;
      let sourceAttributes = attributeSources.current.get(node);
      if (!sourceAttributes) {
        sourceAttributes = new Map();
        attributeSources.current.set(node, sourceAttributes);
      }
      for (const attribute of TRANSLATED_ATTRIBUTES) {
        const current = node.getAttribute(attribute);
        if (current === null) continue;
        if (!sourceAttributes.has(attribute)) sourceAttributes.set(attribute, current);
        const source = sourceAttributes.get(attribute) ?? current;
        const translated = translateText(source, language);
        if (current !== translated) node.setAttribute(attribute, translated);
      }
      for (const child of node.childNodes) translateNode(child);
    };

    const apply = (node: Node) => {
      if (applying) return;
      applying = true;
      translateNode(node);
      applying = false;
    };

    apply(root);
    const observer = new MutationObserver((mutations) => {
      if (applying) return;
      for (const mutation of mutations) {
        if (mutation.type === "characterData") {
          const textNode = mutation.target as Text;
          const source = textSources.current.get(textNode);
          const expected = source === undefined ? undefined : translateText(source, language);
          // A React/state update changed this node: adopt the new English value
          // as its source. Observer callbacks caused by our own translation keep
          // the existing source because the current value already matches.
          if (expected !== undefined && textNode.nodeValue !== expected) {
            textSources.current.set(textNode, textNode.nodeValue ?? "");
          }
          apply(textNode);
        }
        for (const node of mutation.addedNodes) apply(node);
      }
    });
    observer.observe(root, { childList: true, characterData: true, subtree: true });
    return () => observer.disconnect();
  }, [language]);

  return <div ref={rootRef} className="contents">{children}</div>;
}
