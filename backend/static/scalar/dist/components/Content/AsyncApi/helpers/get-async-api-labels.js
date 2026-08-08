import { getAsyncApiServers } from "@scalar/workspace-store/channel-example";
//#region src/components/Content/AsyncApi/helpers/get-async-api-labels.ts
/**
* Server names and protocols a channel is available on.
*
* Resolved from `document.servers`, restricted to `channel.servers` when the channel declares them.
* `webSocketOnly` is disabled so labels cover every protocol, not just WebSocket, and protocols are
* de-duplicated while preserving declaration order.
*/
var getChannelServerLabels = (document, channel) => {
	const entries = getAsyncApiServers(document, {
		channel: channel ?? null,
		webSocketOnly: false
	});
	return {
		servers: entries.map((entry) => entry.name),
		protocols: [...new Set(entries.map((entry) => entry.protocol).filter(Boolean))]
	};
};
//#endregion
export { getChannelServerLabels };

//# sourceMappingURL=get-async-api-labels.js.map