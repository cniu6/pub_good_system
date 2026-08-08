import { useLocalization } from "../../features/localization/use-localization.js";
import { computed, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, openBlock, renderSlot, toDisplayString, unref } from "vue";
import { ScalarIconLink } from "@scalar/icons";
//#region src/components/Anchor/WithBreadcrumb.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["id"];
var _hoisted_2 = { class: "sr-only" };
var WithBreadcrumb_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "WithBreadcrumb",
	props: {
		breadcrumb: {},
		eventBus: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		/** Screen-reader label for the copy-link button, naming the deep-linked item. */
		const copyLinkLabel = computed(() => translate("actions.copyLinkTo", { name: __props.breadcrumb?.[__props.breadcrumb.length - 1] ?? "" }));
		return (_ctx, _cache) => {
			return __props.breadcrumb && __props.breadcrumb.length > 0 ? (openBlock(), createElementBlock("div", {
				key: 0,
				id: __props.breadcrumb.join("."),
				class: "relative scroll-mt-24"
			}, [renderSlot(_ctx.$slots, "default"), createElementVNode("button", {
				class: "text-c-3 hover:text-c-1 absolute -top-2 -left-4.5 flex h-[calc(100%+16px)] w-4.5 cursor-pointer items-center justify-center pr-1.5 opacity-0 group-hover:opacity-100 focus-visible:opacity-100",
				type: "button",
				onClick: _cache[0] || (_cache[0] = () => __props.eventBus?.emit("copy-url:nav-item", { id: __props.breadcrumb.join(".") }))
			}, [createVNode(unref(ScalarIconLink), {
				class: "size-3",
				weight: "bold"
			}), createElementVNode("span", _hoisted_2, [renderSlot(_ctx.$slots, "sr-label", {}, () => [createTextVNode(toDisplayString(copyLinkLabel.value), 1)])])])], 8, _hoisted_1)) : renderSlot(_ctx.$slots, "default", {}, void 0, void 0, 1);
		};
	}
});
//#endregion
export { WithBreadcrumb_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=WithBreadcrumb.vue.script.js.map