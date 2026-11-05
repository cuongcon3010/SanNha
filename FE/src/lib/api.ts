export const apiOrigin =
  process.env.NEXT_PUBLIC_API_ORIGIN || "http://localhost:8080";

export class ApiError extends Error {
  readonly status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
    this.name = "ApiError";
  }
}

export async function api<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const response = await fetch(`${apiOrigin}/api/v1${path}`, {
    ...options,
    credentials: "include",
    cache: "no-store",
    headers: {
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...options.headers,
    },
  });
  const body = await response.json();
  if (!response.ok)
    throw new ApiError(
      typeof body.error === "string"
        ? body.error
        : `Yêu cầu thất bại (HTTP ${response.status}).`,
      response.status,
    );
  return body as T;
}

export type LiveBooking = {
  id: string;
  venueId: string;
  courtId: string;
  date: string;
  hour: number;
  total: number;
  deposit: number;
  status: string;
  depositPaid: boolean;
  subtotal: number;
  discount: number;
  voucherCode: string;
};
export type LivePayment = {
  bookingId: string;
  orderId: string;
  amount: number;
  bankCode: string;
  bankAccount: string;
  bankName: string;
  content: string;
  qrImage?: string;
  status: string;
  bookingStatus: string;
  expiresAt: number;
  expired: boolean;
  environment: "sandbox" | "production";
  subtotal: number;
  discount: number;
  total: number;
  voucherCode: string;
};
export type LiveQuote = {
  subtotal: number;
  discount: number;
  total: number;
  deposit: number;
  voucherCode: string;
};
export type Voucher = {
  code: string;
  kind: "fixed" | "percent";
  value: number;
  minTotal: number;
  maxDiscount: number;
  startsAt: number;
  endsAt: number;
  maxUses: number;
  usedCount: number;
  active: boolean;
  createdAt: number;
};
export const bookingStatus: Record<string, string> = {
  pending: "Chờ cọc",
  confirmed: "Đã xác nhận",
  completed: "Hoàn thành",
  cancelled: "Đã hủy",
  blocked: "Khóa giờ",
};
