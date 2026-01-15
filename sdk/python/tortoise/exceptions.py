"""
Tortoise Exceptions
Custom exceptions for the Tortoise SDK.
"""


class TortoiseError(Exception):
    """Base exception for Tortoise errors."""
    pass


class AuthError(TortoiseError):
    """Authentication error."""
    pass


class ConnectionError(TortoiseError):
    """Connection error."""
    pass


class TimeoutError(TortoiseError):
    """Request timeout error."""
    pass


class ValidationError(TortoiseError):
    """Validation error."""
    pass


class NotFoundError(TortoiseError):
    """Resource not found error."""
    pass


class RateLimitError(TortoiseError):
    """Rate limit exceeded error."""
    pass


class ChannelError(TortoiseError):
    """Channel-related error."""
    pass


class PluginError(TortoiseError):
    """Plugin-related error."""
    pass


class MemoryError(TortoiseError):
    """Memory-related error."""
    pass
