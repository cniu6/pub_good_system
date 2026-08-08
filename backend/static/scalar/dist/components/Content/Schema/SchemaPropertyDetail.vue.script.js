import { createCommentVNode, createElementBlock, createTextVNode, defineComponent, normalizeClass, openBlock, renderSlot } from "vue";
//#region src/components/Content/Schema/SchemaPropertyDetail.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "property-detail-prefix"
};
var _hoisted_2 = {
	key: 1,
	class: "property-detail-value"
};
var _hoisted_3 = {
	key: 2,
	class: "property-detail-value"
};
var SchemaPropertyDetail_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SchemaPropertyDetail",
	props: {
		truncate: { type: Boolean },
		code: { type: Boolean }
	},
	setup(__props) {
		return (_ctx, _cache) => {
			return openBlock(), createElementBlock("span", { class: normalizeClass(["property-detail", { "property-detail-truncate": __props.truncate }]) }, [_ctx.$slots.prefix ? (openBlock(), createElementBlock("div", _hoisted_1, [renderSlot(_ctx.$slots, "prefix", {}, void 0, true), _cache[0] || (_cache[0] = createTextVNode("\xA0 ", -1))])) : createCommentVNode("", true), __props.code ? (openBlock(), createElementBlock("code", _hoisted_2, [renderSlot(_ctx.$slots, "default", {}, void 0, true)])) : (openBlock(), createElementBlock("span", _hoisted_3, [renderSlot(_ctx.$slots, "default", {}, void 0, true)]))], 2);
		};
	}
});
//#endregion
export { SchemaPropertyDetail_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SchemaPropertyDetail.vue.script.js.map