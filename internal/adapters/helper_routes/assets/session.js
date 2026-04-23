/**
 * Normalized session context exposed to the frontend for telemetry and UI decisions.
 * This object is safe for client-side use and does not contain raw LTI claims.
 *
 * @typedef {Object} FrontendSessionContext
 *
 * @property {string} u - Internal user identifier (stable, non-PII ID used for analytics and tracking).
 * @property {string} t - Tenant identifier (e.g., school or organization).
 * @property {Role[]} r - List of normalized roles assigned to the user within the context.
 * @property {string} c - Context identifier (e.g., course, section, or grouping).
 * @property {string} p - Launch platform identifier (e.g., LMS source like Canvas, Schoology, etc.).
 * @property {boolean} i - Indicates whether the session is impersonated (true = acting as another user).
 */

/**
 * Normalized internal role identifier.
 * Values should be controlled and mapped from external systems (e.g., LTI roles).
 *
 * @typedef {string} Role
 */

(() => {
	/**
	 * Initialize telemetry providers (PostHog + Sentry) using frontend session context.
	 *
	 * @param {FrontendSessionContext} t
	 */
	function initTelemetry(t) {
		console.log("Initializing telemetry..");

		window.__LTI_SESSION__ = t;
		window.__LTI_SESSION_INTEGRATIONS__ = [];

		if (t.i) {
			console.warn("Telemetry disabled for impersonated sessions");
			return;
		}

		const ph = typeof window !== "undefined" ? window.posthog : undefined;
		const sentry = typeof window !== "undefined" ? window.Sentry : undefined;

		// PostHog
		if (ph && typeof ph.identify === "function") {
			console.debug("Configuring PostHog session");
			ph.identify(t.u, {
				tenant_id: t.t,
				roles: t.r,
				context_id: t.c,
				platform: t.p,
				impostering: t.i,
			});
			window.__LTI_SESSION_INTEGRATIONS__.push("posthog");
		}

		// Sentry
		if (sentry) {
			console.debug("Configuring Sentry session");

			const scope =
				typeof sentry.getIsolationScope === "function"
					? sentry.getIsolationScope()
					: typeof sentry.getCurrentScope === "function"
						? sentry.getCurrentScope()
						: null;

			if (scope) {
				scope.setUser({ id: t.u });
				scope.setTag("tenant_id", t.t);
				scope.setTag("platform", t.p);
				scope.setTag("impostering", String(t.i));

				if (t.c) scope.setTag("context_id", t.c);
				if (t.r && t.r.length > 0) {
					scope.setTag("roles", t.r.join(","));
				}
			} else {
				if (typeof sentry.setUser === "function") {
					sentry.setUser({ id: t.u });
				}
				if (typeof sentry.setTag === "function") {
					sentry.setTag("tenant_id", t.t);
					sentry.setTag("platform", t.p);
					sentry.setTag("impostering", String(t.i));

					if (t.c) sentry.setTag("context_id", t.c);
					if (t.r && t.r.length > 0) {
						sentry.setTag("roles", t.r.join(","));
					}
				}
			}

			window.__LTI_SESSION_INTEGRATIONS__.push("sentry");
		}
	}
	/** @type {FrontendSessionContext} */
	const payload = {
		u: "{{.UserId}}",
		t: "{{.TenantId}}",
		r: JSON.parse(`{{.RolesJSON}}`),
		c: "{{.ContextId}}",
		p: "{{.LaunchPlatform}}",
		i: JSON.parse(`{{.Impostering}}`),
	};

	initTelemetry(payload);
})();
