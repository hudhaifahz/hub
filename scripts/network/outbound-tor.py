#!/usr/bin/env python3
"""Optional, separate outbound Tor service. Never reads or writes wallet data."""
import argparse
import os
from pathlib import Path
import plistlib
import subprocess
import sys

LABEL = "io.grynvault.tor-outbound"


def prepare(directory, tor_binary):
    directory = directory.resolve()
    if any(char in str(directory) for char in '\n\r"\\'):
        raise ValueError("Unsupported character in configuration path")
    directory.mkdir(parents=True, exist_ok=True, mode=0o700)
    config, agent = directory / "torrc", directory / (LABEL + ".plist")
    if config.exists() or agent.exists():
        raise ValueError("Existing configuration found; refusing to overwrite it")
    (directory / "tor-data").mkdir(mode=0o700)
    config.write_text(
        f'DataDirectory "{directory / "tor-data"}"\n'
        'SocksPort 127.0.0.1:9052\nClientOnly 1\nLog notice stdout\n'
    )
    config.chmod(0o600)
    payload = {
        "Label": LABEL,
        "ProgramArguments": [str(tor_binary), "-f", str(config)],
        "RunAtLoad": True,
        "KeepAlive": True,
        "ThrottleInterval": 30,
        "StandardOutPath": str(directory / "service.log"),
        "StandardErrorPath": str(directory / "service.log"),
    }
    agent.write_bytes(plistlib.dumps(payload))
    agent.chmod(0o600)
    subprocess.run([str(tor_binary), "--verify-config", "-f", str(config)], check=True)
    return agent


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action", choices=["prepare", "install", "status", "uninstall"])
    parser.add_argument("--directory", type=Path, help="Required for prepare; no service is started")
    parser.add_argument("--tor", type=Path, help="Path to an existing Tor executable")
    args = parser.parse_args()
    home = Path.home()
    destination = home / "Library/LaunchAgents" / (LABEL + ".plist")
    service = f"gui/{os.getuid()}/{LABEL}"
    if args.action == "status":
        return subprocess.run(["launchctl", "print", service]).returncode
    if args.action == "uninstall":
        # Retain data and config. Removal never touches an inbound Tor identity.
        if not destination.exists():
            raise ValueError("No Grynvault outbound service is installed")
        active = subprocess.run(["launchctl", "print", service], capture_output=True)
        if active.returncode == 0:
            subprocess.run(["launchctl", "bootout", service], check=True)
        destination.unlink()
        print("Removed outbound service. Configuration and data retained.")
        return 0
    tor = args.tor or next((Path(p) for p in [
        "/opt/homebrew/opt/tor/bin/tor", "/usr/local/opt/tor/bin/tor"
    ] if Path(p).is_file()), None)
    if tor is None or not tor.is_file() or not os.access(tor, os.X_OK):
        raise ValueError("Install Tor first, then supply its executable with --tor")
    if args.action == "prepare":
        if args.directory is None:
            raise ValueError("prepare requires --directory")
        print(f"Prepared and validated: {prepare(args.directory, tor.resolve())}")
        return 0
    if sys.platform != "darwin" or os.getuid() == 0:
        raise ValueError("Install as your normal macOS user")
    if args.directory is not None:
        raise ValueError("--directory is only supported for prepare")
    if destination.exists():
        raise ValueError("An outbound service already exists; use status to inspect it")
    # Refuse a conflicting listener rather than disturbing another service.
    import socket
    with socket.socket() as probe:
        probe.bind(("127.0.0.1", 9052))
    directory = home / "Library/Application Support/Grynvault Network"
    agent = prepare(directory, tor.resolve())
    destination.parent.mkdir(parents=True, exist_ok=True)
    with destination.open("xb") as output:
        output.write(agent.read_bytes())
    destination.chmod(0o600)
    try:
        subprocess.run(["launchctl", "bootstrap", f"gui/{os.getuid()}", str(destination)], check=True)
    except subprocess.CalledProcessError:
        destination.unlink()
        raise
    print("Outbound Tor service installed. Wait for 'Bootstrapped 100%' in service.log.")
    print("Peer routing remains opt-in; follow README.md. No wallet settings were changed.")
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (ValueError, OSError, subprocess.CalledProcessError) as error:
        sys.exit(str(error))
