import { redirectToProxy } from "@scalar/helpers/url/redirect-to-proxy";
//#region src/helpers/upload-temp-document.ts
/** Type guard for the response body */
function isResponseBody(data) {
	return !!data && typeof data === "object" && "url" in data && typeof data.url === "string";
}
/** Upload a document and return a temporary URL */
async function uploadTempDocument(document, urls) {
	const body = JSON.stringify({ document });
	const uploadUrl = `${urls.apiBaseUrl}/core/share/upload/apis`;
	const response = await fetch(redirectToProxy(urls.proxyUrl, uploadUrl), {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body
	});
	if (!response.ok) throw new Error(` Failed to generate temporary link, server responded with ${response.status}`);
	const data = await response.json();
	if (!isResponseBody(data)) throw new Error("Failed to generate temporary link, invalid response from server");
	return data.url;
}
//#endregion
export { uploadTempDocument };

//# sourceMappingURL=upload-temp-document.js.map