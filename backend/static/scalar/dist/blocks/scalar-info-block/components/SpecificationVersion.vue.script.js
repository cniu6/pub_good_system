import Badge_default from "../../../components/Badge/Badge.vue.js";
import { computed, createBlock, createCommentVNode, createTextVNode, defineComponent, openBlock, toDisplayString, unref, withCtx } from "vue";
import { getDocumentTypeLabel } from "@scalar/workspace-store/schemas/type-guards";
//#region src/blocks/scalar-info-block/components/SpecificationVersion.vue?vue&type=script&setup=true&lang.ts
var SpecificationVersion_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "SpecificationVersion",
	props: {
		documentType: { default: "openapi" },
		version: {}
	},
	setup(__props) {
		const label = computed(() => getDocumentTypeLabel(__props.documentType));
		return (_ctx, _cache) => {
			return __props.version ? (openBlock(), createBlock(unref(Badge_default), { key: 0 }, {
				default: withCtx(() => [createTextVNode(toDisplayString(label.value) + " " + toDisplayString(__props.version), 1)]),
				_: 1
			})) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { SpecificationVersion_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=SpecificationVersion.vue.script.js.map