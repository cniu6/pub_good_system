import { useLocalization } from "../../../localization/use-localization.js";
import Callback_default from "./Callback.vue.js";
import { Fragment, computed, createBlock, createElementBlock, createElementVNode, defineComponent, openBlock, renderList, toDisplayString, unref } from "vue";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { isHttpMethod } from "@scalar/helpers/http/is-http-method";
import { objectEntries } from "@scalar/helpers/object/object-entries";
//#region src/features/Operation/components/callbacks/Callbacks.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["aria-label"];
var _hoisted_2 = { class: "callbacks-title text-c-1 my-3 text-lg font-medium" };
/** Extract the callbacks with method, url, and name */
var Callbacks_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Callbacks",
	props: {
		path: {},
		callbacks: {},
		eventBus: {},
		document: {},
		options: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		const flattenedCallbacks = computed(() => {
			const _callbacks = [];
			objectEntries(__props.callbacks).forEach(([name, pathItem]) => {
				objectEntries(getResolvedRef(pathItem)).forEach(([url, methods]) => {
					if (typeof methods !== "object" || !methods) return;
					objectEntries(methods).forEach(([callbackMethod, callback]) => {
						if (!isHttpMethod(callbackMethod)) return;
						_callbacks.push({
							name,
							url,
							method: callbackMethod,
							callback
						});
					});
				});
			});
			return _callbacks;
		});
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", {
				"aria-label": unref(translate)("operation.callbacks"),
				class: "callbacks-list gap-3",
				role: "group"
			}, [createElementVNode("div", _hoisted_2, toDisplayString(unref(translate)("operation.callbacks")), 1), (openBlock(true), createElementBlock(Fragment, null, renderList(flattenedCallbacks.value, ({ callback, method, name, url }) => {
				return openBlock(), createBlock(Callback_default, {
					key: `${name}-${url}-${method}`,
					callback,
					document: __props.document,
					eventBus: __props.eventBus,
					method,
					name,
					options: __props.options,
					path: __props.path,
					url
				}, null, 8, [
					"callback",
					"document",
					"eventBus",
					"method",
					"name",
					"options",
					"path",
					"url"
				]);
			}), 128))], 8, _hoisted_1);
		};
	}
});
//#endregion
export { Callbacks_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Callbacks.vue.script.js.map