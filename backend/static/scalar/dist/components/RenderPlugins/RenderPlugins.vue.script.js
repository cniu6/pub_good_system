import { usePluginManager } from "../../plugins/hooks/usePluginManager.js";
import RenderPluginView_default from "./RenderPluginView.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, defineComponent, openBlock, renderList } from "vue";
//#region src/components/RenderPlugins/RenderPlugins.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "plugin-view"
};
var RenderPlugins_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "RenderPlugins",
	props: {
		viewName: {},
		options: {},
		eventBus: {},
		documentSlug: {}
	},
	setup(__props) {
		const { getViewComponents } = usePluginManager();
		const components = computed(() => getViewComponents(__props.viewName, __props.documentSlug));
		return (_ctx, _cache) => {
			return components.value.length ? (openBlock(), createElementBlock("div", _hoisted_1, [(openBlock(true), createElementBlock(Fragment, null, renderList(components.value, (item) => {
				return openBlock(), createBlock(RenderPluginView_default, {
					key: item.id,
					eventBus: __props.eventBus,
					item,
					options: __props.options
				}, null, 8, [
					"eventBus",
					"item",
					"options"
				]);
			}), 128))])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { RenderPlugins_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=RenderPlugins.vue.script.js.map