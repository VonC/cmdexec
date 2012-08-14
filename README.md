go-cmdexec
==========

**`cmd`** wrapper for real-time logging

Goals:

- follow the stdout/stderr without waiting for the completion of the command
- ignore certain "error" messages (which actually are only warnings)
- detect the exit error (but still returns 'success' if all error messages were ignored)
