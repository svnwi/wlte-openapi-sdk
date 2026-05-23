from typing import Optional


class WlteApiError(Exception):
    def __init__(
        self,
        *,
        status: int,
        code: str,
        message: str,
        data: Optional[object] = None,
        retry_after: Optional[str] = None,
    ) -> None:
        super().__init__(message)
        self.status = status
        self.code = code
        self.message = message
        self.data = data
        self.retry_after = retry_after


def is_auth_expired(error: BaseException) -> bool:
    return isinstance(error, WlteApiError) and error.status == 401 and error.code == "AUTH_EXPIRED"
