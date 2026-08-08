import { useLocalization } from "../../../features/localization/use-localization.js";
import { isFeaturedClient } from "../helpers/featured-clients.js";
import { computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createVNode, defineComponent, normalizeClass, openBlock, ref, toDisplayString, unref, withCtx } from "vue";
import { filterClientsByQuery, findClient } from "@scalar/blocks/code-example";
import { ScalarIcon } from "@scalar/components/icon";
import { ScalarCombobox } from "@scalar/components/combobox";
import { freezeElement } from "@scalar/helpers/dom/freeze-element";
//#region src/blocks/scalar-client-selector-block/components/ClientDropdown.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	"aria-hidden": "true",
	class: "client-libraries-icon__more"
};
var _hoisted_2 = {
	key: 1,
	class: "client-libraries-icon",
	height: "50",
	role: "presentation",
	viewBox: "0 0 50 50",
	width: "50",
	xmlns: "http://www.w3.org/2000/svg"
};
var _hoisted_3 = {
	key: 0,
	class: "client-libraries-text client-libraries-text-more"
};
var _hoisted_4 = { class: "sr-only" };
var ClientDropdown_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ClientDropdown",
	props: {
		clientOptions: {},
		selectedClient: {},
		eventBus: {}
	},
	setup(__props) {
		const containerRef = ref();
		const { translate } = useLocalization();
		/**
		* Icons have longer names to appear in icon searches, e.g. "javascript-js" instead of just "javascript". This function
		* maps the language key to the icon name.
		*/
		const getIconByLanguageKey = (targetKey) => `programming-language-${targetKey === "js" ? "javascript" : targetKey}`;
		/** Set custom example, or update the selected HTTP client globally */
		const selectClient = (option) => {
			if (!containerRef.value) return;
			const unfreeze = freezeElement(containerRef.value);
			setTimeout(() => {
				unfreeze();
			}, 300);
			if (option.clientKey !== "custom") __props.eventBus.emit("workspace:update:selected-client", option.id);
		};
		/** Calculates the targetKey from the selected client id */
		const selectedTargetKey = computed(() => __props.selectedClient?.split("/")[0]);
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", {
				ref_key: "containerRef",
				ref: containerRef,
				class: "client-libraries-more"
			}, [createVNode(unref(ScalarCombobox), {
				filterFn: unref(filterClientsByQuery),
				modelValue: unref(findClient)(__props.clientOptions, __props.selectedClient),
				options: __props.clientOptions,
				placement: "bottom-end",
				teleport: "",
				"onUpdate:modelValue": _cache[0] || (_cache[0] = ($event) => selectClient($event))
			}, {
				default: withCtx(() => [createElementVNode("button", {
					class: normalizeClass(["client-libraries client-libraries__select", { "client-libraries__active": __props.selectedClient && !unref(isFeaturedClient)(__props.selectedClient) }]),
					type: "button"
				}, [
					createElementVNode("div", _hoisted_1, [__props.selectedClient && !unref(isFeaturedClient)(__props.selectedClient) ? (openBlock(), createElementBlock("div", {
						key: 0,
						class: normalizeClass(`client-libraries-icon__${selectedTargetKey.value}`)
					}, [selectedTargetKey.value ? (openBlock(), createBlock(unref(ScalarIcon), {
						key: 0,
						class: "client-libraries-icon",
						icon: getIconByLanguageKey(selectedTargetKey.value)
					}, null, 8, ["icon"])) : createCommentVNode("", true)], 2)) : (openBlock(), createElementBlock("svg", _hoisted_2, [..._cache[1] || (_cache[1] = [createElementVNode("g", {
						fill: "currentColor",
						"fill-rule": "nonzero"
					}, [createElementVNode("path", { d: "M10.71 25.3a3.87 3.87 0 1 0 7.74 0 3.87 3.87 0 0 0-7.74 0M21.13 25.3a3.87 3.87 0 1 0 7.74 0 3.87 3.87 0 0 0-7.74 0M31.55 25.3a3.87 3.87 0 1 0 7.74 0 3.87 3.87 0 0 0-7.74 0" })], -1)])]))]),
					__props.clientOptions.length ? (openBlock(), createElementBlock("span", _hoisted_3, toDisplayString(unref(translate)("clientLibraries.more")), 1)) : createCommentVNode("", true),
					createElementVNode("span", _hoisted_4, toDisplayString(unref(translate)("clientLibraries.selectAll")), 1)
				], 2)]),
				_: 1
			}, 8, [
				"filterFn",
				"modelValue",
				"options"
			])], 512);
		};
	}
});
//#endregion
export { ClientDropdown_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ClientDropdown.vue.script.js.map