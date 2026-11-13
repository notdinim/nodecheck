"""end-to-end install tests for nodecheck

for each distro, it starts a clean container, installs the package that
GoReleaser built into dist/, and checks that the installed tool works

after building packages, go to repo root and run:
    goreleaser release --snapshot --clean
    python3 -m pytest tests/e2e -v
"""
import json
import subprocess
from pathlib import Path

import pytest

DIST = Path(__file__).resolve().parents[2] / "dist"

# (container image, package type, install command, command (checks if package is installed))
DISTROS = [
    ("ubuntu:24.04", "deb", "dpkg -i /pkg/{file}", "dpkg -s nodecheck"),
    ("debian:12", "deb", "dpkg -i /pkg/{file}", "dpkg -s nodecheck"),
    ("rockylinux:9", "rpm", "rpm -i /pkg/{file}", "rpm -q nodecheck"),
]


def find_package(ext):
    """return the amd64 package of the given type from dist/, or skip"""
    matches = sorted(DIST.glob(f"*_linux_amd64.{ext}"))
    if not matches:
        pytest.skip(f"no .{ext} in dist/, run: goreleaser release --snapshot --clean")
    return matches[-1]


def sh(container_id, command):
    
    return subprocess.run(
        ["docker", "exec", container_id, "sh", "-c", command],
        capture_output=True, text=True, timeout=120,
    )


@pytest.fixture(scope="module", params=DISTROS, ids=lambda d: d[0])
def node(request):

    """starts a clean container for one distro, installs the package, and
    removes the container after all tests for that distro have run."""
    image, ext, install_cmd, query_cmd = request.param
    pkg = find_package(ext)

    started = subprocess.run(
        ["docker", "run", "-d", "--rm", "-v", f"{DIST}:/pkg:ro", image, "sleep", "infinity"],
        capture_output=True, text=True, timeout=600,
    )
    if started.returncode != 0:
        pytest.fail(f"could not start {image}:\n{started.stderr}")
    cid = started.stdout.strip()

    try:
        installed = sh(cid, install_cmd.format(file=pkg.name))
        if installed.returncode != 0:
            pytest.fail(f"install failed on {image}:\n{installed.stderr}")
        yield {"id": cid, "image": image, "pkg": pkg, "query": query_cmd}
    finally:
        subprocess.run(["docker", "rm", "-f", cid], capture_output=True)


def test_binary_is_installed(node):
    assert sh(node["id"], "test -x /usr/bin/nodecheck").returncode == 0


def test_package_manager_knows_it(node):
    result = sh(node["id"], node["query"])
    assert result.returncode == 0, result.stderr


def test_version_matches_package(node):
    expected = node["pkg"].name.split("_")[1] # nodecheck_0.2.0_linux_amd64.deb -> "0.2.0"
    result = sh(node["id"], "nodecheck --version")
    assert expected in result.stdout


def test_json_report_makes_sense(node):
    result = sh(node["id"], "nodecheck --json")
    assert result.returncode in (0, 1), result.stderr  # 1 means a warning
    report = json.loads(result.stdout)

    assert report["hostname"]
    assert report["cpus"] > 0
    assert 0 < report["mem_available_mb"] <= report["mem_total_mb"]
    assert report["disk_total_gb"] > 0


def test_uninstall_is_clean(node):
    # testing removing the package on each distro, the binary must be gone

    remove = "dpkg -r nodecheck" if node["pkg"].suffix == ".deb" else "rpm -e nodecheck"
    assert sh(node["id"], remove).returncode == 0
    assert sh(node["id"], "test -e /usr/bin/nodecheck").returncode != 0