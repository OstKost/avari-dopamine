import type { paths } from "./generated-types";

export type ApiPaths = paths;

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export interface RequestOptions extends Omit<RequestInit, "body"> {
  params?: {
    query?: Record<string, string | number | boolean | undefined>;
    path?: Record<string, string | number>;
  };
  body?: unknown;
}

export class ApiError extends Error {
  status: number;
  statusText: string;
  data?: unknown;

  constructor(status: number, statusText: string, message: string, data?: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.statusText = statusText;
    this.data = data;
  }
}

export function isUnauthorizedError(err: unknown): boolean {
  if (err instanceof ApiError && err.status === 401) {
    return true;
  }
  if (err && typeof err === "object" && "status" in err && (err as { status: unknown }).status === 401) {
    return true;
  }
  if (err instanceof Error) {
    const msg = err.message.toLowerCase();
    return (
      msg.includes("401") ||
      msg.includes("unauthorized") ||
      msg.includes("missing access token") ||
      msg.includes("invalid or expired access token")
    );
  }
  return false;
}

export async function apiFetch<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  let url = `${API_BASE_URL}${endpoint}`;

  if (options.params?.path) {
    for (const [key, value] of Object.entries(options.params.path)) {
      url = url.replace(`{${key}}`, encodeURIComponent(String(value)));
    }
  }

  if (options.params?.query) {
    const searchParams = new URLSearchParams();
    for (const [key, value] of Object.entries(options.params.query)) {
      if (value !== undefined && value !== null) {
        searchParams.append(key, String(value));
      }
    }
    const queryString = searchParams.toString();
    if (queryString) {
      url += `?${queryString}`;
    }
  }

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...options.headers,
  };

  const { body, params: _, ...restOptions } = options;
  void _;

  const config: RequestInit = {
    ...restOptions,
    headers,
    credentials: restOptions.credentials || "include",
  };

  if (body !== undefined) {
    config.body = JSON.stringify(body);
  }

  const response = await fetch(url, config);

  if (!response.ok) {
    let errorMsg = `HTTP Error ${response.status}: ${response.statusText}`;
    let errJson: unknown = null;
    try {
      errJson = await response.json();
      if (errJson && typeof errJson === "object") {
        const obj = errJson as Record<string, unknown>;
        if (typeof obj.message === "string") {
          errorMsg = obj.message;
        } else if (typeof obj.error === "string") {
          errorMsg = obj.error;
        }
      }
    } catch {
      // fallback to status text
    }
    throw new ApiError(response.status, response.statusText, errorMsg, errJson);
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return {} as T;
  }

  return response.json() as Promise<T>;
}
