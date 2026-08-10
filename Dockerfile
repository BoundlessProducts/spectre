# ─── Stage 1: builder ───────────────────────────────────────────────────────
FROM golang:1.24-bookworm AS builder

# Install Rust toolchain
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs \
      | sh -s -- -y --default-toolchain stable --profile minimal
ENV PATH="/root/.cargo/bin:${PATH}"

WORKDIR /src
COPY . .

# Build Go binary
RUN go build -trimpath -o /out/spectre ./cmd/spectre

# Build Rust binary
RUN cargo build --manifest-path rust/spectre-mine-rs/Cargo.toml --release \
 && cp rust/spectre-mine-rs/target/release/spectre-mine-rs /out/

# ─── Stage 2: runtime ────────────────────────────────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates python3 \
 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/spectre          /usr/local/bin/spectre
COPY --from=builder /out/spectre-mine-rs  /usr/local/bin/spectre-mine-rs

COPY --from=builder /src/examples   /examples
COPY --from=builder /src/test-suite /test-suite
COPY --from=builder /src/artifact   /artifact
RUN chmod +x /artifact/run-benchmarks.sh

WORKDIR /examples

ENTRYPOINT ["spectre"]
CMD ["--help"]
