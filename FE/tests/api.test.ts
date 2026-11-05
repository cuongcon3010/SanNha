import assert from "node:assert/strict";
import test from "node:test";
import { api, ApiError, apiOrigin } from "../src/lib/api.ts";

test("Backend requests include the session and never cache payment status", async (t) => {
  t.mock.method(globalThis, "fetch", async (url: string, init: RequestInit) => {
    assert.equal(url, `${apiOrigin}/api/v1/bookings/test/payment`);
    assert.equal(init.credentials, "include");
    assert.equal(init.cache, "no-store");
    assert.equal(init.method, "POST");
    assert.equal(init.body, "{}");
    assert.equal(
      (init.headers as Record<string, string>)["Content-Type"],
      "application/json",
    );
    return Response.json({ status: "pending", amount: 66000 });
  });
  const p = await api<{ status: string; amount: number }>(
    "/bookings/test/payment",
    { method: "POST", body: "{}" },
  );
  assert.equal(p.status, "pending");
  assert.equal(p.amount, 66000);
});

test("Unauthorized and disabled payments preserve HTTP status for UI controls", async (t) => {
  for (const status of [401, 403, 404, 503]) {
    const mock = t.mock.method(globalThis, "fetch", async () =>
      Response.json({ error: "Không thể thanh toán." }, { status }),
    );
    await assert.rejects(
      api("/bookings/test/payment"),
      (err: unknown) =>
        err instanceof ApiError &&
        err.status === status &&
        err.message === "Không thể thanh toán.",
    );
    mock.mock.restore();
  }
});
