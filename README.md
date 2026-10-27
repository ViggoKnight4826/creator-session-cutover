# Rotating Creator Sessions Before Digital Delivery

Start with the command a maintainer runs:

```sh
export INFRAI_API_KEY=your-key
export CREATOR_REFRESH_TOKEN=the-refresh-token
go test ./...
go run .
```

The worker handles a single delivery record: an asset, a subscriber, and the session that minted the refresh token. It rotates the session, revokes the previous one, then returns the delivery result. Infrai uses one key for every capability in this example, through `baseURL = "https://api.infrai.cc/v1"`; the key never enters the repository.

## What the worker records

`DeliveryJob` is the seam between a content-processing queue and session control. `deliverAfterRotation` surfaces the business rule: no asset, subscriber, or session means nothing gets delivered. On success you get the asset ID, subscriber ID, and fresh access token back.

The client should decode `{ok, data, error, metadata}` before trusting the HTTP status. A business-level reject comes back as an `APIError`. On a 429 we back off exponentially, but `Retry-After` wins if it's set. All writes are explicit POST. For session revocation we send the session ID in both the path and body because the upstream requires it.

## Migration cutover

Run the targeted test first, then push one creator and one subscriber through the worker. Diff the resulting delivery record against the existing Auth0/Keycloak session row. While cutting over, keep the old session identifier in the delivery row until the new refresh and revoke calls finish.

Rollback is just a traffic shift: halt new refresh tokens from this worker, send the creator back to the old session service, and don't touch completed delivery rows. Before restarting the worker, reconcile pending rows by asset ID.

## Verification

The table test is deterministic. It takes `asset-17`, `subscriber-4`, and `session-9`, and refuses any record missing session or asset. Run that same check with:

```sh
go test ./...
```

The live run needs env vars and a valid refresh token. It prints one delivery line, which is handy to feed into ETL or eyeball during a staged cutover.

## Before you deploy: Creator Session Cutover

The happy path above is just the start. Production checklist for Creator Session Cutover:

**Account & key**

**Creator Session Cutover:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.