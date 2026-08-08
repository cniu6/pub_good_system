import { authStorage, clientStorage } from "../helpers/storage.js";
import { debounce } from "@scalar/helpers/general/debounce";
//#region src/plugins/persistence-plugin.ts
/**
* Plugin to persist workspace state changes with debounced writes.
*/
var persistencePlugin = ({ debounceDelay = 500, maxWait = 1e4, persistAuth = false }) => {
	const { execute } = debounce({
		delay: debounceDelay,
		maxWait
	});
	const authPersistence = authStorage();
	const clientPersistence = clientStorage();
	const getPersistAuth = () => {
		if (typeof persistAuth === "function") return persistAuth();
		return persistAuth;
	};
	return { hooks: { 
	/**
	* Handles all workspace state change events.
	* Each write is debounced by a key to prevent frequent writes for the same entity.
	*/
onWorkspaceStateChanges(event) {
		if (event.type === "meta") {
			const defaultClient = event.value["x-scalar-default-client"];
			if (defaultClient !== void 0) execute("x-scalar-default-client", () => clientPersistence.set(defaultClient));
			return;
		}
		if (getPersistAuth() && event.type === "auth") execute(`auth-${event.documentName}`, () => authPersistence.setAuth(event.documentName, event.value));
	} } };
};
//#endregion
export { persistencePlugin };

//# sourceMappingURL=persistence-plugin.js.map