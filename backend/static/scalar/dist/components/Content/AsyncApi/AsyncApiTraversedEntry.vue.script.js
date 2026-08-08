import Model_default from "../Models/Model.vue.js";
import ModelTag_default from "../Models/ModelTag.vue.js";
import Tag_default from "../Tags/Tag.vue.js";
import Lazy_default from "../../Lazy/Lazy.vue.js";
import { getAsyncApiModelSchema } from "../../../helpers/get-async-api-model-schema.js";
import Channel_default from "./Channel.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createVNode, defineComponent, openBlock, renderList, resolveComponent, unref, withCtx } from "vue";
import { getResolvedRef, mergeSiblingReferences } from "@scalar/workspace-store/helpers/get-resolved-ref";
//#region src/components/Content/AsyncApi/AsyncApiTraversedEntry.vue?vue&type=script&setup=true&lang.ts
var AsyncApiTraversedEntry_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "AsyncApiTraversedEntry",
	props: {
		entries: {},
		document: {},
		expandedItems: {},
		options: {},
		eventBus: {},
		level: { default: 0 }
	},
	setup(__props) {
		const isTagGroup = (entry) => entry.type === "tag" && entry.isGroup === true;
		/**
		* Narrow to a regular (non-group) tag. Tag groups go through a separate branch
		* and must not inflate the sibling-tag count used for `moreThanOneTag`,
		* otherwise `ModernLayout` shows a "Show more" button for a lone tag whenever
		* it sits next to a tag group.
		*/
		const isTag = (entry) => entry.type === "tag" && !isTagGroup(entry);
		const isChannel = (entry) => entry.type === "asyncapi-channel";
		/** The top-level "Models" container that wraps individual schema entries. */
		const isModelsTag = (entry) => entry.type === "models";
		/** A single schema rendered as a model. */
		const isModel = (entry) => entry.type === "model";
		/**
		* Reusable schemas live under `components.schemas`, the same place OpenAPI keeps them.
		* Resolve the wrapper once (merging `$ref` siblings) so the template can gate the Models section.
		*/
		const componentSchemas = computed(() => __props.document.components ? getResolvedRef(__props.document.components, mergeSiblingReferences).schemas : void 0);
		/** Resolve a schema by name into the shape the shared Model component expects. */
		const getModelSchema = (name) => getAsyncApiModelSchema(__props.document, name);
		return (_ctx, _cache) => {
			const _component_AsyncApiTraversedEntry = resolveComponent("AsyncApiTraversedEntry", true);
			return openBlock(true), createElementBlock(Fragment, null, renderList(__props.entries, (entry) => {
				return openBlock(), createBlock(Lazy_default, {
					id: entry.id,
					key: `${entry.id}-${__props.options.layout}`,
					expanded: !!__props.expandedItems[entry.id]
				}, {
					default: withCtx(() => [isChannel(entry) ? (openBlock(), createBlock(Channel_default, {
						key: 0,
						channel: entry,
						document: __props.document,
						eventBus: __props.eventBus,
						expandedItems: __props.expandedItems,
						isCollapsed: !__props.expandedItems[entry.id],
						layout: __props.options.layout,
						level: __props.level,
						options: __props.options
					}, null, 8, [
						"channel",
						"document",
						"eventBus",
						"expandedItems",
						"isCollapsed",
						"layout",
						"level",
						"options"
					])) : isTag(entry) || isTagGroup(entry) && __props.options.layout === "classic" ? (openBlock(), createBlock(unref(Tag_default), {
						key: 1,
						eventBus: __props.eventBus,
						isCollapsed: !__props.expandedItems[entry.id],
						layout: __props.options.layout,
						moreThanOneTag: __props.entries.filter(isTag).length > 1,
						tag: entry
					}, {
						default: withCtx(() => [entry.children?.length ? (openBlock(), createBlock(_component_AsyncApiTraversedEntry, {
							key: 0,
							document: __props.document,
							entries: entry.children,
							eventBus: __props.eventBus,
							expandedItems: __props.expandedItems,
							level: __props.level + 1,
							options: __props.options
						}, null, 8, [
							"document",
							"entries",
							"eventBus",
							"expandedItems",
							"level",
							"options"
						])) : createCommentVNode("", true)]),
						_: 2
					}, 1032, [
						"eventBus",
						"isCollapsed",
						"layout",
						"moreThanOneTag",
						"tag"
					])) : isTagGroup(entry) ? (openBlock(), createBlock(_component_AsyncApiTraversedEntry, {
						key: 2,
						document: __props.document,
						entries: entry.children ?? [],
						eventBus: __props.eventBus,
						expandedItems: __props.expandedItems,
						level: __props.level + 1,
						options: __props.options
					}, null, 8, [
						"document",
						"entries",
						"eventBus",
						"expandedItems",
						"level",
						"options"
					])) : isModelsTag(entry) && componentSchemas.value ? (openBlock(), createBlock(ModelTag_default, {
						key: 3,
						id: entry.id,
						eventBus: __props.eventBus,
						isCollapsed: !__props.expandedItems[entry.id],
						layout: __props.options.layout,
						modelsSectionLabel: __props.options.modelsSectionLabel
					}, {
						default: withCtx(() => [createVNode(_component_AsyncApiTraversedEntry, {
							document: __props.document,
							entries: entry.children ?? [],
							eventBus: __props.eventBus,
							expandedItems: __props.expandedItems,
							level: __props.level + 1,
							options: __props.options
						}, null, 8, [
							"document",
							"entries",
							"eventBus",
							"expandedItems",
							"level",
							"options"
						])]),
						_: 2
					}, 1032, [
						"id",
						"eventBus",
						"isCollapsed",
						"layout",
						"modelsSectionLabel"
					])) : isModel(entry) && getModelSchema(entry.name) ? (openBlock(), createBlock(Model_default, {
						key: 4,
						id: entry.id,
						eventBus: __props.eventBus,
						isCollapsed: !__props.expandedItems[entry.id],
						name: entry.name,
						options: __props.options,
						schema: getModelSchema(entry.name)
					}, null, 8, [
						"id",
						"eventBus",
						"isCollapsed",
						"name",
						"options",
						"schema"
					])) : createCommentVNode("", true)]),
					_: 2
				}, 1032, ["id", "expanded"]);
			}), 128);
		};
	}
});
//#endregion
export { AsyncApiTraversedEntry_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=AsyncApiTraversedEntry.vue.script.js.map