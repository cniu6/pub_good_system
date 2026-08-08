import { useIntersection } from "../../hooks/use-intersection.js";
import { createBlock, createElementBlock, createVNode, defineComponent, mergeProps, normalizeProps, openBlock, resolveDynamicComponent, unref, useTemplateRef, withCtx } from "vue";
import { ScalarErrorBoundary } from "@scalar/components/error-boundary";
//#region src/components/RenderPlugins/RenderPluginView.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = ["id"];
var RenderPluginView_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "RenderPluginView",
	props: {
		item: {},
		options: {},
		eventBus: {}
	},
	setup(__props) {
		const el = useTemplateRef("el");
		/**
		* Participate in the existing scroll-spy. We only emit when the view opts into a sidebar
		* entry, because otherwise we would select a navigation item that does not exist and clear
		* the currently active section as the user scrolls past the plugin view.
		*/
		useIntersection(el, () => {
			if (__props.item.sidebar?.show) __props.eventBus?.emit("intersecting:nav-item", { id: __props.item.id });
		});
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("div", {
				id: __props.item.id,
				ref_key: "el",
				ref: el
			}, [createVNode(unref(ScalarErrorBoundary), null, {
				default: withCtx(() => [__props.item.renderer ? (openBlock(), createBlock(resolveDynamicComponent(__props.item.renderer), normalizeProps(mergeProps({ key: 0 }, {
					component: __props.item.component,
					options: __props.options,
					...__props.item.props
				})), null, 16)) : (openBlock(), createBlock(resolveDynamicComponent(__props.item.component), normalizeProps(mergeProps({ key: 1 }, {
					options: __props.options,
					...__props.item.props
				})), null, 16))]),
				_: 1
			})], 8, _hoisted_1);
		};
	}
});
//#endregion
export { RenderPluginView_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=RenderPluginView.vue.script.js.map