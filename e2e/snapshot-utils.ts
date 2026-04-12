export interface SanitizeOptions {
  maskNewIds?: boolean;
}

/**
 * Sanitizes an API response body for use in snapshot assertions.
 *
 * - Always replaces `http://localhost:<port>` with `http://localhost:TEST_PORT` so that
 *   randomly-assigned ports don't break snapshots.
 * - When `maskNewIds` is true, additionally replaces non-seed entity IDs
 *   (rot_, mem_, usr_, ovr_ prefixes without "SEED" in the value) with
 *   `<<GENERATED_ID>>` and ISO 8601 timestamps with `<<TIMESTAMP>>`, so that
 *   mutation tests (create rotation/member/override) produce stable snapshots.
 */
export function sanitizeApiResponse(
  body: unknown,
  options: SanitizeOptions = {},
): string {
  let json = JSON.stringify(body, null, 2);

  // Always: replace localhost port in link URLs
  json = json.replace(/http:\/\/localhost:\d+/g, "http://localhost:TEST_PORT");

  if (options.maskNewIds) {
    // Replace non-seed entity IDs (those NOT containing "SEED")
    json = json.replace(/\b(rot|mem|usr|ovr)_[A-Za-z0-9]+/g, (match) =>
      match.includes("SEED") ? match : "<<GENERATED_ID>>",
    );

    // Replace ISO 8601 timestamps
    json = json.replace(
      /\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})/g,
      "<<TIMESTAMP>>",
    );
  }

  return json + "\n";
}
