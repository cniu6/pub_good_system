import { useLocalization } from "../features/localization/use-localization.js";
import { Fragment, createBlock, createCommentVNode, createElementBlock, createElementVNode, defineComponent, guardReactiveProps, normalizeClass, normalizeProps, openBlock, renderSlot, toDisplayString, unref } from "vue";
import { ScalarIconList, ScalarIconX } from "@scalar/icons";
import { ScalarIconButton } from "@scalar/components/icon-button";
import { cva } from "@scalar/use-hooks/useBindCx";
//#region src/components/MobileHeader.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = { class: "flex h-(--scalar-header-height) w-full items-center border-b bg-inherit px-2" };
var _hoisted_2 = {
	key: 1,
	class: "flex-1 text-sm font-medium whitespace-nowrap"
};
var _hoisted_3 = { class: "flex h-6 items-center gap-1 pl-1" };
var MobileHeader_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "MobileHeader",
	props: {
		breadcrumb: {},
		isSidebarOpen: { type: Boolean },
		showSidebar: { type: Boolean }
	},
	emits: ["toggleSidebar"],
	setup(__props, { emit: __emit }) {
		const { translate } = useLocalization();
		const emit = __emit;
		const variants = cva({
			base: "lg:hidden items-center bg-b-1 sticky top-(--scalar-custom-header-height,0) z-10 [grid-area:header]",
			variants: { open: { true: "h-(--refs-sidebar-height) custom-scrollbar flex flex-col z-50" } }
		});
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock(Fragment, null, [renderSlot(_ctx.$slots, "sidebar", normalizeProps(guardReactiveProps({ sidebarClasses: "hidden lg:flex sticky top-(--refs-header-height) h-(--refs-sidebar-height) w-(--refs-sidebar-width) [grid-area:navigation]" }))), createElementVNode("div", { class: normalizeClass(["t-doc__header", unref(variants)({ open: __props.isSidebarOpen })]) }, [createElementVNode("header", _hoisted_1, [
				__props.showSidebar ? (openBlock(), createBlock(unref(ScalarIconButton), {
					key: 0,
					icon: __props.isSidebarOpen ? unref(ScalarIconX) : unref(ScalarIconList),
					label: __props.isSidebarOpen ? unref(translate)("navigation.closeMenu") : unref(translate)("navigation.openMenu"),
					size: "md",
					onClick: _cache[0] || (_cache[0] = ($event) => emit("toggleSidebar"))
				}, null, 8, ["icon", "label"])) : createCommentVNode("", true),
				__props.showSidebar ? (openBlock(), createElementBlock("span", _hoisted_2, toDisplayString(__props.breadcrumb), 1)) : renderSlot(_ctx.$slots, "search", {}, void 0, void 0, 2),
				createElementVNode("div", _hoisted_3, [renderSlot(_ctx.$slots, "actions")])
			]), __props.isSidebarOpen ? renderSlot(_ctx.$slots, "sidebar", normalizeProps(guardReactiveProps({ sidebarClasses: "overflow-y-auto custom-scrollbar min-h-0 flex-1 w-full border-none" })), void 0, void 0, 0) : createCommentVNode("", true)], 2)], 64);
		};
	}
});
//#endregion
export { MobileHeader_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=MobileHeader.vue.script.js.map