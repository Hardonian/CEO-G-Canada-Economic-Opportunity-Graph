import type { Metadata } from "next";
import DecisionPlanningWorkbench from "@/components/DecisionPlanningWorkbench";
import { getProjects } from "@/lib/data";

export const metadata: Metadata = {
  title: "Decision Planning Workbench | CanadaOpportunityGraph",
  description:
    "Compare transparent capital-delivery scenarios, portfolio exposure, uncertainty ranges, and sovereign investment priorities across Canadian strategic projects.",
};

export default async function PlanningPage() {
  const projects = await getProjects();
  return <DecisionPlanningWorkbench projects={projects} />;
}
