#!/usr/bin/env sh
set -eu

OPENAPI_FILE="openapi/openapi.yaml"

if [ ! -f "$OPENAPI_FILE" ]; then
  echo "Missing $OPENAPI_FILE" >&2
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required to run OpenAPI Generator without installing local tooling." >&2
  exit 1
fi

GENERATOR_IMAGE="${OPENAPI_GENERATOR_IMAGE:-openapitools/openapi-generator-cli:v7.10.0}"

echo "Validating $OPENAPI_FILE..."
docker run --rm \
  -v "$PWD:/work" \
  "$GENERATOR_IMAGE" validate \
  -i /work/"$OPENAPI_FILE"

echo "Generating TypeScript SDK low-level client..."
docker run --rm \
  -v "$PWD:/work" \
  "$GENERATOR_IMAGE" generate \
  -i /work/"$OPENAPI_FILE" \
  -g typescript-fetch \
  -o /work/sdk/typescript/generated \
  --additional-properties=typescriptThreePlus=true,supportsES6=true

echo "Generating Python SDK low-level client..."
docker run --rm \
  -v "$PWD:/work" \
  "$GENERATOR_IMAGE" generate \
  -i /work/"$OPENAPI_FILE" \
  -g python \
  -o /work/sdk/python/generated \
  --additional-properties=packageName=wlte_openapi_generated

echo "Generating Go SDK low-level client..."
docker run --rm \
  -v "$PWD:/work" \
  "$GENERATOR_IMAGE" generate \
  -i /work/"$OPENAPI_FILE" \
  -g go \
  -o /work/sdk/go/generated \
  --additional-properties=packageName=wlte_openapi_generated

echo "SDK generation completed."
