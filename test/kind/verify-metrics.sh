#! /bin/sh

# Verifies the operator's /metrics endpoint enforces authentication and
# authorization and serves Prometheus metrics to an authorized client.
#
# This guards the controller-runtime built-in metrics auth
# (filters.WithAuthenticationAndAuthorization) that replaced the
# kube-rbac-proxy sidecar. The deploy smoke test only checks that the
# operator pod is healthy, so without this a broken metrics auth path
# would go unnoticed.

set -eu

KUBECTL_BIN=${KUBECTL_BIN:-kubectl}
NAMESPACE=${NAMESPACE:-shipwright-operator}
SERVICE_ACCOUNT=${SERVICE_ACCOUNT:-shipwright-operator}
METRICS_SERVICE=${METRICS_SERVICE:-shipwright-operator-metrics-service}
METRICS_READER_ROLE=${METRICS_READER_ROLE:-shipwright-operator-metrics-reader}
METRICS_PORT=${METRICS_PORT:-8443}
LOCAL_PORT=${LOCAL_PORT:-8443}
CRB_NAME="shipwright-operator-metrics-verify"

PF_PID=""

cleanup() {
    if [ -n "${PF_PID}" ]; then
        kill "${PF_PID}" >/dev/null 2>&1 || true
    fi
    ${KUBECTL_BIN} delete clusterrolebinding "${CRB_NAME}" --ignore-not-found >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "# Granting ${NAMESPACE}/${SERVICE_ACCOUNT} read access to /metrics via ${METRICS_READER_ROLE}"
${KUBECTL_BIN} create clusterrolebinding "${CRB_NAME}" \
    --clusterrole="${METRICS_READER_ROLE}" \
    --serviceaccount="${NAMESPACE}:${SERVICE_ACCOUNT}" \
    --dry-run=client -o yaml | ${KUBECTL_BIN} apply -f -

echo "# Minting a token for ${NAMESPACE}/${SERVICE_ACCOUNT}"
TOKEN=$(${KUBECTL_BIN} create token "${SERVICE_ACCOUNT}" -n "${NAMESPACE}")

echo "# Port-forwarding svc/${METRICS_SERVICE} ${LOCAL_PORT}:${METRICS_PORT}"
${KUBECTL_BIN} port-forward -n "${NAMESPACE}" "svc/${METRICS_SERVICE}" "${LOCAL_PORT}:${METRICS_PORT}" >/dev/null 2>&1 &
PF_PID=$!

echo "# Waiting for the metrics endpoint to become reachable"
ready=""
i=1
while [ "${i}" -le 30 ]; do
    # curl exits 0 for any HTTP response (including 401), non-zero only when
    # the tunnel is not yet accepting connections.
    if curl -sk -o /dev/null "https://localhost:${LOCAL_PORT}/metrics"; then
        ready="true"
        break
    fi
    sleep 2
    i=$((i + 1))
done
if [ -z "${ready}" ]; then
    echo "ERROR: metrics endpoint did not become reachable"
    exit 1
fi

echo "# Unauthenticated request should be rejected with HTTP 401"
unauth_code=$(curl -sk -o /dev/null -w '%{http_code}' "https://localhost:${LOCAL_PORT}/metrics")
if [ "${unauth_code}" != "401" ]; then
    echo "ERROR: expected HTTP 401 without a token, got ${unauth_code}"
    exit 1
fi
echo "  OK: got HTTP 401 without a token"

echo "# Authenticated request should return Prometheus metrics"
auth_body=$(curl -sk -H "Authorization: Bearer ${TOKEN}" "https://localhost:${LOCAL_PORT}/metrics")
if ! echo "${auth_body}" | grep -q '^# HELP'; then
    echo "ERROR: expected Prometheus metrics with a valid token, got:"
    echo "${auth_body}" | head
    exit 1
fi
echo "  OK: received Prometheus metrics with a valid token"

echo "# Metrics endpoint authentication and authorization verified"
