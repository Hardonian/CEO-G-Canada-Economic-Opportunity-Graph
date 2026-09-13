"""
COG SDK Exceptions
"""

class COGError(Exception):
    """Base exception for COG SDK errors."""
    def __init__(self, message: str, status_code: int = None, response_body: dict = None):
        super().__init__(message)
        self.status_code = status_code
        self.response_body = response_body


class COGNotFoundError(COGError):
    """Resource not found (404)."""
    def __init__(self, message: str = "Resource not found", **kwargs):
        super().__init__(message, status_code=404, **kwargs)


class COGValidationError(COGError):
    """Request validation error (400)."""
    def __init__(self, message: str = "Validation error", **kwargs):
        super().__init__(message, status_code=400, **kwargs)


class COGRateLimitError(COGError):
    """Rate limit exceeded (429)."""
    def __init__(self, message: str = "Rate limit exceeded", retry_after: int = None, **kwargs):
        super().__init__(message, status_code=429, **kwargs)
        self.retry_after = retry_after


class COGServerError(COGError):
    """Server error (5xx)."""
    def __init__(self, message: str = "Server error", **kwargs):
        super().__init__(message, status_code=500, **kwargs)