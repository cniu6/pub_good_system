import { useLocalization } from "../../../features/localization/use-localization.js";
import ScreenReader_default from "../../ScreenReader.vue.js";
import { scrollTargetId } from "../../../helpers/lazy-bus.js";
import { SCHEMA_DYNAMIC_SCOPE_SYMBOL, resolveDynamicSchema, useDynamicScope } from "./helpers/dynamic-scope.js";
import { inferDiscriminatorMappingComposition } from "./helpers/get-compositions-to-render.js";
import { isTypeObject } from "./helpers/is-type-object.js";
import { isEmptySchemaObject } from "./helpers/is-empty-schema-object.js";
import { mergeAllOfSchemas } from "./helpers/merge-all-of-schemas.js";
import { SCHEMA_ANCESTORS_SYMBOL } from "./helpers/schema-cycle.js";
import SchemaComposition_default from "./SchemaComposition.vue.js";
import SchemaHeading_default from "./SchemaHeading.vue.js";
import SchemaProperty_default from "./SchemaProperty.vue.js";
import SchemaObjectProperties_default from "./SchemaObjectProperties.vue.js";
import { Fragment, computed, createBlock, createCommentVNode, createElementBlock, createElementVNode, createTextVNode, createVNode, defineComponent, inject, normalizeClass, normalizeStyle, openBlock, provide, toDisplayString, unref, vShow, withCtx, withDirectives } from "vue";
import { Disclosure, DisclosureButton, DisclosurePanel } from "@headlessui/vue";
import { ScalarMarkdown } from "@scalar/components/markdown";
import { ScalarIcon } from "@scalar/components/icon";
import { resolve } from "@scalar/workspace-store/resolve";
import { pushDynamicScope } from "@scalar/workspace-store/helpers/dynamic-ref";
//#region src/components/Content/Schema/Schema.vue?vue&type=script&setup=true&lang.ts
var _hoisted_1 = {
	key: 0,
	class: "schema-card-description"
};
var _hoisted_2 = {
	key: 1,
	class: "pt-2"
};
var _hoisted_3 = {
	key: 0,
	class: "schema-properties"
};
var Schema_vue_vue_type_script_setup_true_lang_default = /*@__PURE__*/ defineComponent({
	__name: "Schema",
	props: {
		schema: {},
		level: { default: 0 },
		name: {},
		compact: { type: Boolean },
		noncollapsible: {
			type: Boolean,
			default: false
		},
		hideHeading: { type: Boolean },
		hideDescription: {
			type: Boolean,
			default: false
		},
		additionalProperties: { type: Boolean },
		hideModelNames: {
			type: Boolean,
			default: false
		},
		discriminator: {},
		breadcrumb: {},
		eventBus: {},
		options: {},
		schemaContext: {},
		compositionPath: {},
		cycleKey: {}
	},
	setup(__props) {
		const { translate } = useLocalization();
		/**
		* The dynamic scope inherited from ancestor schema resources.
		*
		* Used to bind JSON Schema 2020-12 `$dynamicRef`s to the active `$dynamicAnchor` while walking the
		* tree. Empty at the root. See {@link useDynamicScope}.
		*/
		const dynamicScope = useDynamicScope();
		/**
		* The schema this node actually renders.
		*
		* Two normalizations happen here, both no-ops for ordinary schemas:
		* - A top-level `$dynamicRef` is bound to its concrete type via the inherited dynamic scope.
		* - A resource that extends a template through a root `$ref` (JSON Schema 2020-12 `$ref` alongside
		*   `$defs`, e.g. a `PaginatedResponse` binding) is merged so its inherited properties render.
		*/
		const resolvedSchema = computed(() => {
			if (!__props.schema || typeof __props.schema !== "object") return __props.schema;
			const bound = resolveDynamicSchema(__props.schema, dynamicScope);
			return "$ref" in bound ? resolve.schema(bound) : bound;
		});
		/**
		* Re-provide the dynamic scope grown with this resource so nested `$dynamicRef`s bind here.
		*
		* Built once at setup from the resource's stable identity (like the ancestor set below);
		* `pushDynamicScope` only grows the scope for schemas that can carry a `$dynamicAnchor`.
		*
		* The raw schema is pushed, not the merged {@link resolvedSchema}: merging through `resolve.schema`
		* coerces the node and drops the resolved `$ref-value` from entries inside `$defs`, which
		* `$dynamicAnchor` resolution relies on to dereference the bound type (e.g. `User`).
		*/
		const scopeSchema = __props.schema ? resolveDynamicSchema(__props.schema, dynamicScope) : void 0;
		provide(SCHEMA_DYNAMIC_SCOPE_SYMBOL, scopeSchema ? pushDynamicScope(dynamicScope, scopeSchema) : dynamicScope);
		/**
		* Cycle-safe `expandAllSchemaProperties`.
		*
		* We track ancestor schema keys along the current render path. A node is
		* treated as cyclic when its key is already present in the ancestor set, which
		* indicates that rendering has looped back onto a self-referential schema.
		*
		* This lets us default-expand finite branches while stopping automatic
		* expansion only at cycle boundaries, preventing infinite recursion.
		*/
		const ancestors = inject(SCHEMA_ANCESTORS_SYMBOL, void 0);
		const isCyclic = computed(() => __props.cycleKey != null && !!ancestors?.has(__props.cycleKey));
		const childAncestors = new Set(ancestors ?? []);
		if (__props.cycleKey != null) childAncestors.add(__props.cycleKey);
		provide(SCHEMA_ANCESTORS_SYMBOL, childAncestors);
		const shouldForceExpand = computed(() => !!__props.options.expandAllSchemaProperties && !isCyclic.value);
		/**
		* Determines whether to show the collapse/expand toggle button.
		* We hide the toggle for non-collapsible schemas and root-level schemas.
		*/
		const shouldShowToggle = computed(() => !__props.noncollapsible && __props.level > 0);
		/**
		* Whether this schema sits on the path to the current anchor/scroll target.
		*
		* Property anchors are dot-joined breadcrumbs, so every disclosure that wraps
		* the target has a breadcrumb that is a prefix of the target id. Opening those
		* disclosures is what makes deep links to collapsed (hidden) properties work
		* without forcing every schema open via `expandAllSchemaProperties`.
		*/
		const isOnScrollTargetPath = computed(() => {
			if (!__props.breadcrumb?.length) return false;
			const path = __props.breadcrumb.join(".");
			const target = scrollTargetId.value;
			return target === path || target.startsWith(`${path}.`);
		});
		/**
		* Whether the disclosure starts expanded. Non-collapsible schemas are always
		* open. When `expandAllSchemaProperties` is enabled, finite branches start
		* expanded by default while cyclic branches remain collapsed to avoid recursion
		* loops. We also open any disclosure on the path to the current scroll target so
		* deep links resolve even when the property is collapsed.
		*
		* Note: the Disclosure only reads this at mount, so it expands collapsed
		* properties on a fresh navigation (the schema mounts after the target is set),
		* not when the target changes for an already-mounted disclosure.
		*/
		const defaultOpen = computed(() => __props.noncollapsible || shouldForceExpand.value || isOnScrollTargetPath.value);
		const childAttributesLabel = computed(() => __props.schema?.title ?? translate("schema.childAttributes"));
		/** Gets the description to show for the schema */
		const schemaDescription = computed(() => {
			const value = resolvedSchema.value;
			if (__props.hideDescription) return null;
			if (__props.schema?.allOf && __props.schema.allOf.length > 0 && __props.name === "Request Body") return mergeAllOfSchemas(__props.schema)?.description || null;
			if (!value?.description || typeof value.description !== "string") return null;
			if (value.enum) return null;
			if (!("properties" in value) && !("patternProperties" in value) && !("additionalProperties" in value)) return null;
			return value.description;
		});
		/**
		* Infer a selector for mapped discriminators that do not declare `oneOf`.
		* Threaded discriminators skip inference to avoid recursive allOf variants.
		*/
		const inferredDiscriminatorComposition = computed(() => __props.schema && !__props.discriminator && isTypeObject(__props.schema) ? inferDiscriminatorMappingComposition(__props.schema, __props.options.document) : null);
		const handleClick = (e) => {
			if (__props.noncollapsible) e.stopPropagation();
		};
		return (_ctx, _cache) => {
			return resolvedSchema.value && Object.keys(resolvedSchema.value).length ? (openBlock(), createBlock(unref(Disclosure), {
				key: 0,
				defaultOpen: defaultOpen.value
			}, {
				default: withCtx(({ open }) => [createElementVNode("div", { class: normalizeClass(["schema-card", [
					`schema-card--level-${__props.level}`,
					{
						"schema-card--compact": __props.compact,
						"schema-card--open": open
					},
					{ "border-t": __props.additionalProperties && open }
				]]) }, [
					schemaDescription.value ? (openBlock(), createElementBlock("div", _hoisted_1, [createVNode(unref(ScalarMarkdown), { value: schemaDescription.value }, null, 8, ["value"])])) : createCommentVNode("", true),
					unref(isEmptySchemaObject)(resolvedSchema.value) ? (openBlock(), createElementBlock("div", _hoisted_2, toDisplayString(unref(translate)("schema.emptyObject")), 1)) : createCommentVNode("", true),
					createElementVNode("div", { class: normalizeClass(["schema-properties", { "schema-properties-open": open }]) }, [__props.additionalProperties ? withDirectives((openBlock(), createElementBlock("div", _hoisted_3, [createVNode(unref(DisclosureButton), {
						as: "button",
						class: "schema-card-title schema-card-title--compact",
						onClickCapture: handleClick
					}, {
						default: withCtx(() => [
							createVNode(unref(ScalarIcon), {
								class: "schema-card-title-icon",
								icon: "Add",
								size: "sm"
							}),
							createTextVNode(" " + toDisplayString(unref(translate)("schema.showAdditionalProperties")) + " ", 1),
							__props.name ? (openBlock(), createBlock(ScreenReader_default, { key: 0 }, {
								default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("schema.forName", { name: __props.name })), 1)]),
								_: 1
							})) : createCommentVNode("", true)
						]),
						_: 1
					})], 512)), [[vShow, !open]]) : shouldShowToggle.value ? withDirectives((openBlock(), createBlock(unref(DisclosureButton), {
						key: 1,
						as: __props.noncollapsible ? "div" : "button",
						class: normalizeClass(["schema-card-title", { "schema-card-title--compact": __props.compact }]),
						style: normalizeStyle({ top: `calc(var(--refs-viewport-offset) +  calc(var(--schema-title-height) * ${__props.level}))` }),
						onClickCapture: handleClick
					}, {
						default: withCtx(() => [__props.compact ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [
							createVNode(unref(ScalarIcon), {
								class: normalizeClass(["schema-card-title-icon", { "schema-card-title-icon--open": open }]),
								icon: "Add",
								size: "sm"
							}, null, 8, ["class"]),
							open ? (openBlock(), createElementBlock(Fragment, { key: 0 }, [createTextVNode(toDisplayString(unref(translate)("schema.hideChildAttributes", { name: childAttributesLabel.value })), 1)], 64)) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [createTextVNode(toDisplayString(unref(translate)("schema.showChildAttributes", { name: childAttributesLabel.value })), 1)], 64)),
							__props.name ? (openBlock(), createBlock(ScreenReader_default, { key: 2 }, {
								default: withCtx(() => [createTextVNode(toDisplayString(unref(translate)("schema.forName", { name: __props.name })), 1)]),
								_: 1
							})) : createCommentVNode("", true)
						], 64)) : (openBlock(), createElementBlock(Fragment, { key: 1 }, [createVNode(unref(ScalarIcon), {
							class: normalizeClass(["schema-card-title-icon", { "schema-card-title-icon--open": open }]),
							icon: "Add",
							size: "sm"
						}, null, 8, ["class"]), createVNode(SchemaHeading_default, {
							name: resolvedSchema.value?.title ?? __props.name,
							value: resolvedSchema.value
						}, null, 8, ["name", "value"])], 64))]),
						_: 2
					}, 1032, [
						"as",
						"class",
						"style"
					])), [[vShow, !__props.hideHeading && !(__props.noncollapsible && __props.compact)]]) : createCommentVNode("", true), !__props.additionalProperties || open ? (openBlock(), createBlock(unref(DisclosurePanel), {
						key: 2,
						as: "ul",
						static: !shouldShowToggle.value
					}, {
						default: withCtx(() => [inferredDiscriminatorComposition.value ? (openBlock(), createBlock(SchemaComposition_default, {
							key: 0,
							breadcrumb: __props.breadcrumb,
							compact: __props.compact,
							composition: "oneOf",
							compositionPath: __props.compositionPath,
							discriminator: __props.schema?.discriminator,
							eventBus: __props.eventBus,
							hideHeading: __props.hideHeading,
							hideModelNames: __props.hideModelNames,
							level: __props.level,
							name: __props.name,
							options: __props.options,
							schema: inferredDiscriminatorComposition.value,
							schemaContext: __props.schemaContext
						}, null, 8, [
							"breadcrumb",
							"compact",
							"compositionPath",
							"discriminator",
							"eventBus",
							"hideHeading",
							"hideModelNames",
							"level",
							"name",
							"options",
							"schema",
							"schemaContext"
						])) : unref(isTypeObject)(resolvedSchema.value) ? (openBlock(), createBlock(SchemaObjectProperties_default, {
							key: 1,
							breadcrumb: __props.breadcrumb,
							compact: __props.compact,
							compositionPath: __props.compositionPath,
							discriminator: __props.discriminator,
							eventBus: __props.eventBus,
							hideHeading: __props.hideHeading,
							hideModelNames: __props.hideModelNames,
							level: __props.level + 1,
							options: __props.options,
							schema: resolvedSchema.value,
							schemaContext: __props.schemaContext
						}, null, 8, [
							"breadcrumb",
							"compact",
							"compositionPath",
							"discriminator",
							"eventBus",
							"hideHeading",
							"hideModelNames",
							"level",
							"options",
							"schema",
							"schemaContext"
						])) : (openBlock(), createElementBlock(Fragment, { key: 2 }, [resolvedSchema.value ? (openBlock(), createBlock(SchemaProperty_default, {
							key: 0,
							breadcrumb: __props.breadcrumb,
							compact: __props.compact,
							compositionPath: __props.compositionPath,
							discriminator: __props.discriminator,
							eventBus: __props.eventBus,
							hideHeading: __props.hideHeading,
							hideModelNames: __props.hideModelNames,
							level: __props.level,
							options: __props.options,
							schema: resolvedSchema.value,
							schemaContext: __props.schemaContext
						}, null, 8, [
							"breadcrumb",
							"compact",
							"compositionPath",
							"discriminator",
							"eventBus",
							"hideHeading",
							"hideModelNames",
							"level",
							"options",
							"schema",
							"schemaContext"
						])) : createCommentVNode("", true)], 64))]),
						_: 1
					}, 8, ["static"])) : createCommentVNode("", true)], 2)
				], 2)]),
				_: 1
			}, 8, ["defaultOpen"])) : createCommentVNode("", true);
		};
	}
});
//#endregion
export { Schema_vue_vue_type_script_setup_true_lang_default as default };

//# sourceMappingURL=Schema.vue.script.js.map