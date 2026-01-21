from typing import Tuple, List

ALLOWED_COMMANDS = ['get', 'describe', 'logs', 'top']
FORBIDDEN_FLAGS = ['-o jsonpath', '-o=jsonpath', '--dry-run', '--overrides']

def validate_kubectl_command(command: str) -> Tuple[bool, str]:
    """
    Validate if a kubectl command is safe to execute.
    Only allows read-only commands.
    """
    parts = command.split()
    if not parts:
        return False, "Empty command"

    cmd_verb = parts[0]
    if cmd_verb not in ALLOWED_COMMANDS:
        return False, f"Command '{cmd_verb}' is not allowed. Only read-only commands are permitted."

    # Check for forbidden flags
    for part in parts:
        if part in FORBIDDEN_FLAGS:
             return False, f"Flag '{part}' is not allowed."

    # Ensure no semicolons or piping to prevent shell injection (basic check)
    if ';' in command or '|' in command or '>' in command or '&' in command:
        return False, "Shell operators are not allowed."

    return True, "Command is valid"
