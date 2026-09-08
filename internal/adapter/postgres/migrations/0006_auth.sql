-- 0006_auth.sql — who is allowed to call the API.
--
-- One table, because there is no identity context yet (P9-PLAN §5). A row is
-- a bearer token and the subject it stands for; the subject's name, role and
-- login live nowhere, because nothing needs them yet.
--
-- token_hash is the PRIMARY KEY, and the token itself is never stored. A dump
-- of this table gives an attacker hashes of random 128-bit strings: nothing to
-- reverse, nothing to reuse. Verify hashes what it is given and looks it up.
CREATE TABLE IF NOT EXISTS api_tokens (
    token_hash text PRIMARY KEY,

    -- 'operator' or 'customer'. Text, not an enum type, for the same reason
    -- every other status column here is text: a psql session should be
    -- readable without the Go code. auth.Kind validates on the way out.
    kind       text        NOT NULL,

    -- The principal's id, as the domain knows it: an operator id for staff, a
    -- customer id for a customer. No foreign key — there is no users table,
    -- and inventing one to satisfy a constraint would be a table nobody reads.
    subject    uuid        NOT NULL,

    -- What this token is for, so a human can revoke the right one later.
    label      text        NOT NULL DEFAULT '',

    created_at timestamptz NOT NULL,

    -- Revocation is a timestamp, not a DELETE: "this token stopped working on
    -- Tuesday" is an answer an audit needs, and a deleted row cannot give it.
    revoked_at timestamptz
);

-- The operator screen lists tokens by subject; the login path never uses this.
CREATE INDEX IF NOT EXISTS api_tokens_subject_idx ON api_tokens (subject);
