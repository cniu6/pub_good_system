import { useLocalization } from "../../localization/use-localization.js";
import Header_default from "./Header.vue.js";
import { Fragment, createBlock, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, normalizeClass, openBlock, renderList, toDisplayString, unref, withCtx } from "vue";
import { Disclosure, DisclosureButton, DisclosurePanel } from "@headlessui/vue";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
import { ScalarIcon } from "@scalar/components/icon";
//#region src/features/Operation/components/Headers.vue?vue&type=script&setup=true&lang.ts
var Headers_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Headers",
	props: {
		headers: {},
		breadcrumb: {},
		eventBus: {},
		document: {},
		orderRequiredPropertiesFirst: { type: Boolean },
		orderSchemaPropertiesBy: {},
		expandAllSchemaProperties: { type: Boolean }
	},
	setup(__props) {
		const { translate } = useLocalization();
		return (_ctx, _cache) => {
			return openBlock(), createBlock(unref(Disclosure), null, {
				default: withCtx(({ open }) => [createElementVNode("div", { class: normalizeClass(["headers-card headers-card--compact", [{ "headers-card--open": open }]]) }, [createElementVNode("div", { class: normalizeClass(["headers-properties", { "headers-properties-open": open }]) }, [createVNode(unref(DisclosureButton), {
					class: "headers-card-title headers-card-title--compact",
					style: { top: `calc(var(--refs-viewport-offset)))` }
				}, {
					default: withCtx(() => [createVNode(unref(ScalarIcon), {
						class: normalizeClass(["headers-card-title-icon", { "headers-card-title-icon--open": open }]),
						icon: "Add",
						size: "sm"
					}, null, 8, ["class"]), open ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [createTextVNode(toDisplayString(unref(translate)("operation.hideHeaders")), 1)], 64)) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [createTextVNode(toDisplayString(unref(translate)("operation.showHeaders")), 1)], 64))]),
					_: 2
				}, 1024), createVNode(unref(DisclosurePanel), null, {
					default: withCtx(() => [(openBlock(true), createElementBlock(Fragment, null, renderList(__props.headers, (header, key) => {
						return openBlock(), createBlock(Header_default, {
							key,
							breadcrumb: __props.breadcrumb ? [...__props.breadcrumb, "headers"] : void 0,
							document: __props.document,
							eventBus: __props.eventBus,
							expandAllSchemaProperties: __props.expandAllSchemaProperties,
							header: unref(getResolvedRef)(header),
							name: key,
							orderRequiredPropertiesFirst: __props.orderRequiredPropertiesFirst,
							orderSchemaPropertiesBy: __props.orderSchemaPropertiesBy
						}, null, 8, [
							"breadcrumb",
							"document",
							"eventBus",
							"expandAllSchemaProperties",
							"header",
							"name",
							"orderRequiredPropertiesFirst",
							"orderSchemaPropertiesBy"
						]);
					}), 128))]),
					_: 1
				})], 2)], 2)]),
				_: 1
			});
		};
	}
});
//#endregion
export { Headers_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Headers.vue.script.js.map