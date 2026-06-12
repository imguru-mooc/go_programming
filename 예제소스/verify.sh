#!/usr/bin/env bash
set -e
echo "=== 실행형 예제 ==="
for d in p1_client_ctx p2_marshal p2_tags p2_dynamic p2_money p2_stream \
         p2_encodings p4_ctxlog p4_recover p5_trace p6_slice; do
  echo "--- $d ---"; (cd "$d" && go run .)
done

echo "=== 테스트형 예제 ==="
for d in p2_money p2_handler p3_mock p3_httptest p3_calc p5_sort p6_openssl; do
  echo "--- $d ---"; (cd "$d" && go test -race ./...)
done

echo "=== race 데모 (검출이 정상) ==="
(cd p4_race && go run -race . 2>&1 | grep -c "DATA RACE" >/dev/null && echo "race 검출 OK")

echo "✅ 전체 검증 완료"
