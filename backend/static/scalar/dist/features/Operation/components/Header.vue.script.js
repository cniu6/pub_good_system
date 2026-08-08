import SchemaProperty_default from "../../../components/Content/Schema/SchemaProperty.vue.js";
import { createBlock, createCommentVNode, defineComponent, openBlock, unref } from "vue";
import { getResolvedRef } from "@scalar/workspace-store/helpers/get-resolved-ref";
//#region src/features/Operation/components/Header.vue?vue&type=script&setup=true&lang.ts
var Header_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Header",
	props: {
		header: {},
		name: {},
		breadcrumb: {},
		eventBus: {},
		document: {},
		orderSchemaPropertiesBy: {},
		orderRequiredPropertiesFirst: { type: Boolean },
		expandAllSchemaProperties: { type: Boolean }
	},
	setup(__props) {
		return (_ctx, _cache) => {
			return "schema" in __props.header && __props.header.schema ? (openBlock(), createBlock(SchemaProperty_default, {
				key: 0,
				breadcrumb: __props.breadcrumb ? [...__props.breadcrumb, "headers"] : void 0,
				description: __props.header.description,
				eventBus: __props.eventBus,
				name: __props.name,
				options: {
					orderRequiredPropertiesFirst: __props.orderRequiredPropertiesFirst,
					orderSchemaPropertiesBy: __props.orderSchemaPropertiesBy,
					expandAllSchemaProperties: __props.expandAllSchemaProperties,
					document: __props.document
				},
				schema: unref(getResolvedRef)(__props.header.schema)
			}, null, 8, [
				"breadcrumb",
				"description",
				"eventBus",
				"name",
				"options",
				"schema"
			])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { Header_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Header.vue.script.js.map