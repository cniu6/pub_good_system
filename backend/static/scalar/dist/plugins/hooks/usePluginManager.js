import { createPluginManager } from "../plugin-manager.js";
import { inject } from "vue";
//#region src/plugins/hooks/usePluginManager.ts
var PLUGIN_MANAGER_SYMBOL = Symbol();
/**
* Hook to access the plugin manager
*/
var usePluginManager = () => {
	const manager = inject(PLUGIN_MANAGER_SYMBOL);
	if (!manager) return createPluginManager({});
	return manager;
};
//#endregion
export { PLUGIN_MANAGER_SYMBOL, usePluginManager };

//# sourceMappingURL=usePluginManager.js.map