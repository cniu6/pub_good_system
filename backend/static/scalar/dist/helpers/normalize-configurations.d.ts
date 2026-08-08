import { type AnyApiReferenceConfiguration, type ApiReferenceConfigurationRaw, type ApiReferenceConfigurationWithSource } from '@scalar/types/api-reference';
/** Processed API Reference Configuration
 *
 * Creates the required title and slug for the API Reference.
 * Separate the source into a dedicated object
 * Returns the raw configuration to pass to components
 */
export type NormalizedConfiguration = {
    title: string;
    slug: string;
    config: ApiReferenceConfigurationRaw;
    default: boolean;
    agent: ApiReferenceConfigurationWithSource['agent'];
    source: {
        url: string;
        content?: never;
    } | {
        content: Record<string, unknown>;
        url?: never;
    };
};
type NormalizedConfigurations = Record<string, NormalizedConfiguration>;
/**
 * Take any configuration and return a flat array of configurations.
 */
export declare const normalizeConfigurations: (configuration: AnyApiReferenceConfiguration | undefined) => NormalizedConfigurations;
/** Normalize content into a JS object or return null if it is falsey */
export declare const normalizeContent: (content: string | Record<string, unknown> | (() => string | Record<string, unknown>)) => Record<string, unknown> | null;
export {};
//# sourceMappingURL=normalize-configurations.d.ts.map