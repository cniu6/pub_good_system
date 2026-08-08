import { apiReferenceConfigurationWithSourceSchema } from "@scalar/schemas/api-reference";
import { isConfigurationWithSources } from "@scalar/types/api-reference";
import { slugger } from "@scalar/helpers/string/slugger";
import { parseJsonOrYaml } from "@scalar/oas-utils/helpers";
//#region src/helpers/normalize-configurations.ts
var isConfigWithRequiredSource = (input) => {
	return !!input.url?.trim() || !!input.content;
};
/**
* Take any configuration and return a flat array of configurations.
*/
var normalizeConfigurations = (configuration) => {
	const { slug } = slugger();
	const normalized = {};
	if (!configuration) return normalized;
	(Array.isArray(configuration) ? configuration : [configuration]).flatMap((c) => {
		if (isConfigurationWithSources(c)) {
			const { sources: configSources, ...rest } = c;
			return configSources?.map((source) => ({
				...rest,
				...source
			})) ?? [];
		}
		return [c];
	}).map((source) => apiReferenceConfigurationWithSourceSchema(source)).filter(isConfigWithRequiredSource).map((source, index) => addSlugAndTitle(source, index, slug)).forEach((c) => {
		const { url, content, ...config } = c;
		normalized[c.slug] = {
			config,
			title: c.title,
			slug: c.slug,
			default: !!c?.default,
			agent: c.agent,
			source: content ? { content: normalizeContent(content) ?? {} } : { url }
		};
	});
	return normalized;
};
/** Normalize content into a JS object or return null if it is falsey */
var normalizeContent = (content) => {
	if (!content) return null;
	if (typeof content === "function") return normalizeContent(content());
	if (typeof content === "string") return parseJsonOrYaml(content);
	return content;
};
/** Process a single spec configuration so that it has a title and a slug */
var addSlugAndTitle = (source, index = 0, slug) => {
	if (source.title) return {
		...source,
		slug: source.slug || slug(source.title),
		title: source.title
	};
	if (source.slug) return {
		...source,
		slug: slug(source.slug),
		title: source.slug
	};
	return {
		...source,
		slug: `api-${index + 1}`,
		title: `API #${index + 1}`
	};
};
//#endregion
export { normalizeConfigurations, normalizeContent };

//# sourceMappingURL=normalize-configurations.js.map