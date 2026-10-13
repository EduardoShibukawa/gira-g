# syntax=docker/dockerfile:1.10@sha256:865e5dd094beca432e8c0a1d5e1c465db5f998dca4e439981029b3b81fb39ed5
# Arm D: Understand Anything, built from a pinned commit (ADR 0012 rule 1).
# Egonex-AI/Understand-Anything main at ffa2f0a84a85525623e9f94f7a6b560e39ae43ce (plugin 2.9.7, MIT).
# Its knowledge graph is produced by an LLM pass inside a host agent: that indexing
# run is a paid, protected action and needs its own owner approval with a cost estimate.
FROM node:22-slim@sha256:c3de60bf2f9dd0ac6370e6117950ff62d6e339527e7472301c9c78a017978392
ADD --keep-git-dir=false https://github.com/Egonex-AI/Understand-Anything.git#ffa2f0a84a85525623e9f94f7a6b560e39ae43ce /src
WORKDIR /src/understand-anything-plugin
# The lockfile pins every JavaScript dependency.
RUN corepack enable && pnpm install --frozen-lockfile
USER node
WORKDIR /corpus
