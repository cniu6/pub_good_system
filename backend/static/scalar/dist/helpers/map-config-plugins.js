import { watch } from "vue";
import { buildRequest, getEnvironmentVariables } from "@scalar/workspace-store/request-example";
import { buildSafeBodyRequest } from "@scalar/helpers/http/can-method-have-body";
//#region src/helpers/map-config-plugins.ts
/**
* Maps API reference configuration callbacks to client plugins.
*
* This function transforms the onBeforeRequest, onRequestBuilt, and onRequestSent
* callbacks into the new plugin hook system. The mapping is reactive, so changes
* to the configuration will automatically update the plugin hooks.
*
* Note: onRequestBuilt is mapped to the requestBuilt hook so the callback receives
* the exact fetch Request that is sent over the wire. Mutating its headers modifies
* the outgoing request, and hashing its body produces a hash that matches what the
* server receives (a rebuilt multipart body would get a different boundary).
*
* Note: onRequestSent is mapped to responseReceived hook. This is not a perfect
* one-to-one mapping, but it maintains backward compatibility with the old API.
* The old callback receives only the URL string, while the new hook receives
* the full response object.
*
* @param config - Reactive configuration object containing optional hook callbacks
* @returns Array containing a single plugin with the mapped hooks
*/
var mapConfigPlugins = (config, environment) => {
	const plugin = { hooks: {} };
	watch([
		() => config.value.onBeforeRequest,
		() => config.value.onRequestBuilt,
		() => config.value.onRequestSent,
		() => environment.value
	], ([onBeforeRequest, onRequestBuilt, onRequestSent, environment]) => {
		const envVariables = getEnvironmentVariables(environment);
		if (!plugin.hooks) plugin.hooks = {};
		plugin.hooks.beforeRequest = onBeforeRequest ? async (payload) => {
			const built = buildRequest(payload.requestBuilder, {
				envVariables,
				allowMissingRequestServerBase: true
			});
			if (!built.ok) {
				console.error("[@scalar/api-reference] onBeforeRequest was not run because the request could not be built:", built.message ?? built.error);
				return;
			}
			await onBeforeRequest({
				request: buildSafeBodyRequest(...built.data.requestPayload),
				requestBuilder: payload.requestBuilder,
				envVariables
			});
		} : void 0;
		/**
		* Maps onRequestBuilt to the requestBuilt hook. The payload request is the
		* exact fetch Request that will be sent, so header mutations apply and body
		* hashes match what goes over the wire.
		*/
		plugin.hooks.requestBuilt = onRequestBuilt ? async (payload) => {
			await onRequestBuilt({
				request: payload.request,
				requestBuilder: payload.requestBuilder,
				envVariables
			});
		} : void 0;
		/**
		* Maps onRequestSent to responseReceived hook.
		* The old API only passed the URL string, so we extract it from the request.
		*/
		plugin.hooks.responseReceived = onRequestSent ? (payload) => {
			onRequestSent(payload.request.url);
		} : void 0;
	}, { immediate: true });
	return [plugin];
};
//#endregion
export { mapConfigPlugins };

//# sourceMappingURL=map-config-plugins.js.map