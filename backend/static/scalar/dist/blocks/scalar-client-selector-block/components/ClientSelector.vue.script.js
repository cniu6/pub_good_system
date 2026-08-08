import { useLocalization } from "../../../features/localization/use-localization.js";
import { getFeaturedClients, isFeaturedClient } from "../helpers/featured-clients.js";
import ClientDropdown_default from "./ClientDropdown.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, normalizeClass, normalizeStyle, openBlock, ref, renderList, toDisplayString, unref, useId, useTemplateRef, watch, withCtx } from "vue";
import { Tab, TabGroup, TabList, TabPanel, TabPanels } from "@headlessui/vue";
import { DEFAULT_CLIENT } from "@scalar/blocks/code-example";
import { ScalarIcon } from "@scalar/components/icon";
//#region src/blocks/scalar-client-selector-block/components/ClientSelector.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	ref: "wrapper-ref"
};
var _hoisted_2 = ["id"];
var _hoisted_3 = { class: "client-libraries-list" };
var _hoisted_4 = { class: "client-libraries-text" };
var _hoisted_5 = ["id"];
var ClientSelector_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "ClientSelector",
	props: {
		clientOptions: {},
		selectedClient: { default: () => DEFAULT_CLIENT },
		eventBus: {}
	},
	setup(__props, { expose: __expose }) {
		const headingId = useId();
		const morePanel = useId();
		const { translate } = useLocalization();
		/**
		* Whether a selection is a custom code sample (e.g. `custom/python`) rather than
		* a built-in client. Custom samples are matched by the `custom/` id prefix, which
		* mirrors the `^custom/` pattern enforced on the stored default client.
		*/
		const isCustomSelection = (client) => Boolean(client?.startsWith("custom/"));
		/**
		* The generic client this selector actually displays.
		*
		* The introduction selector only represents the built-in HTTP clients. Custom
		* code samples are operation-specific and "always just have the generic
		* clients", so when one is selected globally we keep showing the last generic
		* client here instead of switching to (and failing to render) a custom sample.
		*/
		const activeClient = ref(isCustomSelection(__props.selectedClient) ? DEFAULT_CLIENT : __props.selectedClient);
		watch(() => __props.selectedClient, (newClient) => {
			if (!isCustomSelection(newClient)) activeClient.value = newClient;
		});
		/** Grab the option for the currently selected Http Client */
		const selectedClientOption = computed(() => __props.clientOptions.flatMap((optionGroup) => optionGroup.options.find((option) => option.id === activeClient.value) ?? [])[0]);
		/** List of featured clients */
		const featuredClients = computed(() => getFeaturedClients(__props.clientOptions));
		/** Currently selected tab index */
		const tabIndex = computed(() => featuredClients.value.findIndex((client) => client.id === activeClient.value));
		const wrapper = useTemplateRef("wrapper-ref");
		const getIconByLanguageKey = (targetKey) => `programming-language-${targetKey === "js" ? "javascript" : targetKey}`;
		/** Handle tab selection */
		const onTabSelect = (index) => {
			const client = featuredClients.value[index];
			if (!client || !wrapper.value) return;
			__props.eventBus.emit("workspace:update:selected-client", client.id);
		};
		__expose({ selectedClientOption });
		return (_ctx, _cache) => {
			return __props.clientOptions.length ? (openBlock(), createElementBlock("div", _hoisted_1, [createVNode(unref(TabGroup), {
				manual: "",
				selectedIndex: tabIndex.value,
				onChange: onTabSelect
			}, {
				default: withCtx(() => [
					createElementVNode("div", {
						id: unref(headingId),
						class: "client-libraries-heading"
					}, toDisplayString(unref(translate)("clientLibraries.heading")), 9, _hoisted_2),
					createElementVNode("div", _hoisted_3, [createVNode(unref(TabList), {
						"aria-labelledby": unref(headingId),
						class: "client-libraries-tabs",
						style: normalizeStyle({ flexGrow: featuredClients.value.length })
					}, {
						default: withCtx(() => [(openBlock(true), createElementBlock(Fragment, null, renderList(featuredClients.value, (featuredClient) => {
							return openBlock(), createBlock(unref(Tab), {
								key: featuredClient.clientKey,
								class: normalizeClass(["client-libraries rendered-code-sdks", { "client-libraries__active": featuredClient.id === activeClient.value }])
							}, {
								default: withCtx(() => [createElementVNode("div", { class: normalizeClass(`client-libraries-icon__${featuredClient.targetKey}`) }, [createVNode(unref(ScalarIcon), {
									class: "client-libraries-icon",
									icon: getIconByLanguageKey(featuredClient.targetKey)
								}, null, 8, ["icon"])], 2), createElementVNode("span", _hoisted_4, toDisplayString(featuredClient.targetTitle), 1)]),
								_: 2
							}, 1032, ["class"]);
						}), 128))]),
						_: 1
					}, 8, ["aria-labelledby", "style"]), createVNode(ClientDropdown_default, {
						clientOptions: __props.clientOptions,
						eventBus: __props.eventBus,
						selectedClient: activeClient.value
					}, null, 8, [
						"clientOptions",
						"eventBus",
						"selectedClient"
					])]),
					createVNode(unref(TabPanels), null, {
						default: withCtx(() => [unref(isFeaturedClient)(activeClient.value) ? (openBlock(true), createElementBlock(Fragment, { key: 0 }, renderList(featuredClients.value, (client) => {
							return openBlock(), createBlock(unref(TabPanel), {
								key: client.id,
								class: "selected-client card-footer -outline-offset-2"
							}, {
								default: withCtx(() => [createTextVNode(toDisplayString(client.title), 1)]),
								_: 2
							}, 1024);
						}), 128)) : (openBlock(), createElementBlock("div", {
							key: 1,
							id: unref(morePanel),
							class: "selected-client card-footer -outline-offset-2",
							role: "tabpanel",
							tabindex: "0"
						}, toDisplayString(selectedClientOption.value?.title), 9, _hoisted_5))]),
						_: 1
					})
				]),
				_: 1
			}, 8, ["selectedIndex"])], 512)) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { ClientSelector_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=ClientSelector.vue.script.js.map