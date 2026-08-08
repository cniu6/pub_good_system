import type { WorkspaceEventBus } from '@scalar/workspace-store/events';
import type { DiscriminatorObject, SchemaObject } from '@scalar/workspace-store/schemas/v3.1/strict/openapi-document';
import type { SchemaOptions } from '../../../components/Content/Schema/types';
type __VLS_Props = {
    schema: SchemaObject;
    discriminator?: DiscriminatorObject;
    compact?: boolean;
    hideHeading?: boolean;
    level?: number;
    hideModelNames?: boolean;
    breadcrumb?: string[];
    eventBus: WorkspaceEventBus | null;
    options: SchemaOptions;
    schemaContext?: string;
    compositionPath?: string[];
};
declare const __VLS_export: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
declare const _default: typeof __VLS_export;
export default _default;
//# sourceMappingURL=SchemaObjectProperties.vue.d.ts.map