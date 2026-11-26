export interface ProviderConsole {
  id: string;
  name: string;
  url: string;
}

interface ProviderPreset {
  id: string;
  name: string;
  consoleUrl: string;
  keywords: string[];
}

// 依据订阅名称/网站识别常见服务商，给出控制台入口。
// 这是无需改数据库结构的 v1 方案；后续如需显式选择服务商，可以把 id 落到订阅字段里。
const providerPresets: ProviderPreset[] = [
  { id: "aliyun", name: "阿里云", consoleUrl: "https://console.aliyun.com/", keywords: ["aliyun", "阿里云"] },
  { id: "tencent", name: "腾讯云", consoleUrl: "https://console.cloud.tencent.com/", keywords: ["tencent", "腾讯云", "qcloud"] },
  { id: "huawei", name: "华为云", consoleUrl: "https://console.huaweicloud.com/", keywords: ["huaweicloud", "华为云"] },
  { id: "aws", name: "AWS", consoleUrl: "https://console.aws.amazon.com/", keywords: ["aws", "amazon web services"] },
  { id: "gcp", name: "Google Cloud", consoleUrl: "https://console.cloud.google.com/", keywords: ["google cloud", "gcp"] },
  { id: "oracle", name: "Oracle Cloud", consoleUrl: "https://cloud.oracle.com/", keywords: ["oracle cloud", "oci"] },
  { id: "cloudflare", name: "Cloudflare", consoleUrl: "https://dash.cloudflare.com/", keywords: ["cloudflare"] },
  { id: "vultr", name: "Vultr", consoleUrl: "https://my.vultr.com/", keywords: ["vultr"] },
  { id: "digitalocean", name: "DigitalOcean", consoleUrl: "https://cloud.digitalocean.com/", keywords: ["digitalocean", "do "] },
  { id: "hetzner", name: "Hetzner", consoleUrl: "https://console.hetzner.cloud/", keywords: ["hetzner"] },
  { id: "linode", name: "Linode", consoleUrl: "https://cloud.linode.com/", keywords: ["linode", "akamai cloud"] },
  { id: "namecheap", name: "Namecheap", consoleUrl: "https://www.namecheap.com/myaccount/", keywords: ["namecheap"] },
  { id: "godaddy", name: "GoDaddy", consoleUrl: "https://sso.godaddy.com/", keywords: ["godaddy"] },
  { id: "porkbun", name: "Porkbun", consoleUrl: "https://porkbun.com/account", keywords: ["porkbun"] },
  { id: "westcn", name: "西部数码", consoleUrl: "https://www.west.cn/", keywords: ["西部数码", "west.cn"] },
];

export function findProviderConsole(name: string, website: string): ProviderConsole | null {
  const haystack = `${name} ${website}`.toLowerCase();
  for (const preset of providerPresets) {
    if (preset.keywords.some((keyword) => haystack.includes(keyword))) {
      return { id: preset.id, name: preset.name, url: preset.consoleUrl };
    }
  }
  return null;
}
