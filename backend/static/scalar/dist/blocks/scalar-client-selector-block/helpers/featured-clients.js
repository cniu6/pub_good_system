//#region src/blocks/scalar-client-selector-block/helpers/featured-clients.ts
/** Hard coded list of default featured clients */
var FEATURED_CLIENTS = [
	"shell/curl",
	"ruby/native",
	"node/undici",
	"php/guzzle",
	"python/python3"
];
/** Whether or not a client is in the featured list */
var isFeaturedClient = (clientId, featuredClients = FEATURED_CLIENTS) => Boolean(clientId && featuredClients.includes(clientId));
/**
* Maps featured client IDs to their corresponding ClientOption objects.
* Returns an array of ClientOption objects that match the featured clients list,
* maintaining the order of the featured clients.
*/
var getFeaturedClients = (clientOptions, featuredClients = FEATURED_CLIENTS) => {
	const clientMap = /* @__PURE__ */ new Map();
	for (const group of clientOptions) for (const option of group.options) clientMap.set(option.id, option);
	return featuredClients.flatMap((clientId) => {
		return clientMap.get(clientId) ?? [];
	});
};
//#endregion
export { getFeaturedClients, isFeaturedClient };

//# sourceMappingURL=featured-clients.js.map