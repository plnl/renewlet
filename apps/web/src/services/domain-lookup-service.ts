import { z } from "zod";
import { apiSuccessResponseSchema } from "@renewlet/shared/schemas/api";
import { apiFetch } from "@/lib/api-client";

// 域名到期查询服务层：浏览器只调用自家后端，RDAP 请求与代理都由 Go 运行面处理。
const domainLookupPayloadSchema = z.object({
  domain: z.string(),
  expiryDate: z.string(),
  registrar: z.string().optional(),
  source: z.string(),
}).strict();

const domainLookupResponseSchema = apiSuccessResponseSchema(domainLookupPayloadSchema);

export interface DomainLookupResult {
  domain: string;
  expiryDate: string;
  registrar?: string | undefined;
  source: string;
}

export const domainLookupService = {
  async lookup(domain: string): Promise<DomainLookupResult> {
    return await apiFetch("/api/app/subscriptions/domain-lookup", domainLookupResponseSchema, {
      method: "POST",
      body: JSON.stringify({ domain }),
      timeoutMs: 20_000,
    });
  },
};
