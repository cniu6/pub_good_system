import { authStorage, clientStorage } from "./storage.js";
import { isSelectedClient } from "@scalar/blocks/code-example";
//#region src/helpers/load-from-perssistance.ts
/**
* Loads the default HTTP client from storage and applies it to the workspace.
* Only updates if no default client is already set. Accepts both built-in client
* ids and custom sample ids (e.g. `custom/python`) so a custom selection persists.
*/
var loadClientFromStorage = (store) => {
	const storedClient = clientStorage().get();
	if (isSelectedClient(storedClient) && !store.workspace["x-scalar-default-client"]) store.update("x-scalar-default-client", storedClient);
};
/**
* Loads the authentication data from storage and applies it to the workspace.
* Only updates if no authentication data is already set.
*/
var loadAuthFromStorage = (store, slug) => {
	const auth = authStorage().getAuth(slug);
	store.auth.load({ [slug]: auth });
};
//#endregion
export { loadAuthFromStorage, loadClientFromStorage };

//# sourceMappingURL=load-from-perssistance.js.map