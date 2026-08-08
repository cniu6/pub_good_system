import { computed, createBlock, createTextVNode, defineComponent, normalizeStyle, openBlock, renderSlot, resolveDynamicComponent, toDisplayString, withCtx } from "vue";
import { getHttpMethodInfo } from "@scalar/helpers/http/http-info";
import { normalizeHttpMethod } from "@scalar/helpers/http/normalize-http-method";
//#region src/components/HttpMethod/HttpMethod.vue?vue&type=script&setup=true&lang.ts
var HttpMethod_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "HttpMethod",
	props: {
		as: {},
		property: {},
		short: { type: Boolean },
		method: {}
	},
	setup(__props) {
		const props = __props;
		/** Grabs the method info object which contains abbreviation, color, and background color etc */
		const httpMethodInfo = computed(() => getHttpMethodInfo(String(props.method || "")));
		/** Full method name */
		const normalized = computed(() => normalizeHttpMethod(props.method));
		return (_ctx, _cache) => {
			return openBlock(), createBlock(resolveDynamicComponent(__props.as ?? "span"), {
				class: "uppercase",
				style: normalizeStyle({ [__props.property || "color"]: httpMethodInfo.value.colorVar })
			}, {
				default: withCtx(() => [renderSlot(_ctx.$slots, "default"), createTextVNode(" " + toDisplayString(__props.short ? httpMethodInfo.value.short : normalized.value), 1)]),
				_: 3
			}, 8, ["style"]);
		};
	}
});
//#endregion
export { HttpMethod_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=HttpMethod.vue.script.js.map