# CALL-E Hackathon — Milestone 1: Proof of Call

## Goal

Prove the complete path before building the fleet workflow:

```text
TypeScript
   ↓
@call-e/calle
   ↓
One explicit outbound call
   ↓
Structured result
```

CALL-E's official integration repository documents the TypeScript SDK as `@call-e/calle` and the `CalleClient.calls.createAndWait()` flow.

## Run

```bash
cd calle
npm install
export CALLE_API_KEY="calle_live_key"
export CALLE_TEST_PHONE="+E164_TEST_NUMBER"
npm run dev
```

The script makes one real outbound call. **Only use a number you control or have explicit permission to call.** Do not commit API keys or real phone numbers.

## Expected output

```json
{
  "status": "completed",
  "taskCompleted": true,
  "structuredResult": {
    "confirmed": true,
    "summary": "Recipient confirmed they heard the message."
  }
}
```

The exact returned status/result depends on the live call.

## Why this milestone matters

This is intentionally small. Before adding fleet data, dashboards, webhooks, or databases, we need to establish that the application can:

1. authenticate to CALL-E;
2. create an outbound task from TypeScript;
3. complete a real phone call; and
4. receive structured machine-readable output.

Next milestone: turn the proof into a fleet operations agent that can call a driver/dispatcher and capture operational status.
