import { computed, toValue, watch } from "vue";
import { useFavicon } from "@vueuse/core";
import { isClient } from "@scalar/blocks/code-example";
import { useSeoMeta } from "@unhead/vue";
//#region src/helpers/map-config-to-workspace-store.ts
var mapConfigToWorkspaceStore = ({ config, store, isDarkMode }) => {
	watch(() => toValue(config).defaultHttpClient, (newValue) => {
		if (newValue) {
			const { targetKey, clientKey } = newValue;
			const clientId = `${targetKey}/${clientKey}`;
			if (isClient(clientId)) store.update("x-scalar-default-client", clientId);
		}
	}, { immediate: true });
	/** Update the dark mode state when props change */
	watch(() => toValue(config).darkMode, (isDark) => store.update("x-scalar-color-mode", isDark ? "dark" : "light"));
	watch(() => isDarkMode.value, (newIsDark) => store.update("x-scalar-color-mode", newIsDark ? "dark" : "light"), { immediate: true });
	if (toValue(config).metaData) useSeoMeta(toValue(config).metaData);
	watch(() => toValue(config).proxyUrl, (newProxyUrl) => store.update("x-scalar-active-proxy", newProxyUrl), { immediate: true });
	useFavicon(computed(() => toValue(config).favicon));
};
//#endregion
export { mapConfigToWorkspaceStore };

//# sourceMappingURL=map-config-to-workspace-store.js.map