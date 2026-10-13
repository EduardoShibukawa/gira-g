# syntax=docker/dockerfile:1.10@sha256:865e5dd094beca432e8c0a1d5e1c465db5f998dca4e439981029b3b81fb39ed5
# Arm C: Graphify, built from pinned source (ADR 0012 rule 1).
# Graphify-Labs/graphify tag v0.9.82 = commit 5b74d7d74911cf435c8f1636b6f96ea202cc6246 (Apache-2.0).
FROM python:3.14-slim@sha256:a2b82f3c48559aa0a8446d9af49826b6e2b2016f4cd2afabfe6013ec53729170
ADD --keep-git-dir=false https://github.com/Graphify-Labs/graphify.git#5b74d7d74911cf435c8f1636b6f96ea202cc6246 /src
# Python dependencies resolve from PyPI during the build phase only; `pip freeze` is
# written into the image so every report can record the resolved versions.
RUN pip install --no-cache-dir "/src[mcp]" && pip freeze > /pip-freeze.txt
USER nobody
WORKDIR /corpus
