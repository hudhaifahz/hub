// e.g. https://api.boltz.exchange -> boltz.exchange
export function getSwapProviderName(swapApiUrl: string) {
  try {
    const host = new URL(swapApiUrl).host;
    if (host === "swap.coinos.io") {
      return "Coinos";
    }
    return host.replace(/^api\./, "");
  } catch {
    return swapApiUrl;
  }
}
