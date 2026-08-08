import ApiReference_default from "../../components/ApiReference.vue.js";
import { hasPluginUrls, loadPluginsFromUrls } from "./load-plugins-from-urls.js";
import { createApp, createSSRApp, h, reactive } from "vue";
import "@scalar/schemas/api-reference";
import { createHead } from "@unhead/vue/client";
//#region src/standalone/lib/html-api.ts
/**
* The id given to the standalone build's single injected `<style>` tag.
* Keep in sync with `vite.standalone.config.ts` and `vite.standalone.esm.config.ts`.
*/
var STANDALONE_STYLE_ID = "scalar-style";
/**
* Per-document bookkeeping for the standalone build's injected styles.
*
* The CDN build injects all of its CSS into one `<style>` tag in `<head>`. Under
* SPA-style navigation (Turbo Drive, htmx boost, Astro view transitions) the host
* swaps the DOM without reloading the window, so those document-level styles
* (`@layer scalar-base`, the `:root` theme variables) would otherwise linger and
* bleed into the host app's next page. We reference-count the live instances and
* detach the styles when the last one is destroyed, re-attaching them when a new
* instance mounts so navigating back to the reference is still styled.
*
* State is keyed by document so the counter survives navigations (the JS context
* persists) while staying isolated per page.
*/
var standaloneStyleState = /* @__PURE__ */ new WeakMap();
var getStandaloneStyleState = (doc) => {
	const existing = standaloneStyleState.get(doc);
	if (existing) return existing;
	const state = {
		count: 0,
		detachedStyle: null
	};
	standaloneStyleState.set(doc, state);
	return state;
};
/** Track a freshly mounted instance and restore previously detached styles. */
var retainStandaloneStyles = (doc) => {
	const state = getStandaloneStyleState(doc);
	state.count += 1;
	if (state.detachedStyle && !doc.getElementById(STANDALONE_STYLE_ID)) {
		doc.head.appendChild(state.detachedStyle);
		state.detachedStyle = null;
	}
};
/** Release an instance and detach the injected styles once the last one is gone. */
var releaseStandaloneStyles = (doc) => {
	const state = getStandaloneStyleState(doc);
	state.count = Math.max(0, state.count - 1);
	if (state.count > 0) return;
	const styleElement = doc.getElementById(STANDALONE_STYLE_ID);
	if (styleElement instanceof HTMLStyleElement) {
		state.detachedStyle = styleElement;
		styleElement.remove();
	}
};
/**
* Create (and mount) a new Scalar API Reference
*
* @example createApiReference({ url: '/scalar.json' }).mount('#app')
* @example createApiReference('#app', { url: '/scalar.json' })
* @example createApiReference(document.getElementById('app'), { url: '/scalar.json' })
*/
var createApiReference = (elementOrSelectorOrConfig, optionalConfiguration) => {
	const idPrefix = "scalar-refs";
	const props = reactive({ configuration: optionalConfiguration ?? elementOrSelectorOrConfig ?? {} });
	const createReferenceApp = (isSsr = false) => {
		const referenceApp = isSsr ? createSSRApp(() => h(ApiReference_default, props)) : createApp(() => h(ApiReference_default, props));
		referenceApp.use(createHead());
		referenceApp.config.idPrefix = idPrefix;
		return referenceApp;
	};
	const mountElement = optionalConfiguration ? typeof elementOrSelectorOrConfig === "string" ? document.querySelector(elementOrSelectorOrConfig) : elementOrSelectorOrConfig : null;
	let app = createReferenceApp(!!optionalConfiguration && !!mountElement && mountElement.children.length > 0);
	let hasMounted = false;
	if (optionalConfiguration) if (mountElement) {
		const mount = () => {
			app.mount(mountElement);
			hasMounted = true;
			retainStandaloneStyles(document);
		};
		if (hasPluginUrls(props.configuration)) loadPluginsFromUrls(props.configuration).then(() => {
			if (!abortController.signal.aborted) mount();
		});
		else mount();
	} else console.error("Could not find a mount point for API References:", elementOrSelectorOrConfig);
	const abortController = new AbortController();
	const listenerOptions = {
		capture: false,
		signal: abortController.signal
	};
	/**
	* Reload the API Reference
	* @deprecated
	*/
	document.addEventListener("scalar:reload-references", () => {
		console.warn("scalar:reload-references event has been deprecated, please use the scalarInstance.app.mount method instead.");
		if (!props.configuration) return;
		const currentElement = typeof elementOrSelectorOrConfig === "string" ? document.querySelector(elementOrSelectorOrConfig) : elementOrSelectorOrConfig;
		if (!currentElement) return;
		if (currentElement && !document.body.contains(currentElement)) document.body.appendChild(currentElement);
		app.unmount();
		app = createReferenceApp();
		app.mount(currentElement);
	}, listenerOptions);
	/** Destroy the current API Reference instance */
	const destroy = () => {
		abortController.abort();
		props.configuration = {};
		if (hasMounted) {
			hasMounted = false;
			app.unmount();
			releaseStandaloneStyles(document);
		}
	};
	/**
	* Allow user to destroy the API Reference
	* @deprecated
	*/
	document.addEventListener("scalar:destroy-references", () => {
		console.warn("scalar:destroy-references event has been deprecated, please use scalarInstance.destroy instead.");
		destroy();
	}, listenerOptions);
	/**
	* Allow user to update configuration
	* @deprecated
	*/
	document.addEventListener("scalar:update-references-config", (ev) => {
		console.warn("scalar:update-references-config event has been deprecated, please use scalarInstance.updateConfiguration instead.");
		if ("detail" in ev) Object.assign(props, ev.detail);
	}, listenerOptions);
	return {
		app,
		getConfiguration: () => props.configuration ?? {},
		updateConfiguration: (newConfig) => {
			props.configuration = newConfig;
		},
		destroy
	};
};
//#endregion
export { createApiReference };

//# sourceMappingURL=html-api.js.map