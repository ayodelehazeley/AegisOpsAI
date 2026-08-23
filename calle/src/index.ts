import { CalleClient } from "@call-e/calle";

const to = process.env.CALLE_TEST_PHONE;
const apiKey = process.env.CALLE_API_KEY;

if (!apiKey) throw new Error("Missing CALLE_API_KEY");
if (!to) throw new Error("Missing CALLE_TEST_PHONE (E.164 format)");

const client = new CalleClient({ apiKey });

const call = await client.calls.createAndWait({
  task: `Call ${to} and introduce yourself as the CALL-E proof agent. Explain that this is a test call for a fleet operations automation prototype. Ask the recipient to confirm that they heard the message. Return a structured result indicating whether the recipient confirmed.`,
  resultSchema: {
    type: "object",
    required: ["confirmed", "summary"],
    properties: {
      confirmed: { type: "boolean" },
      summary: { type: "string" },
    },
  },
});

console.log(JSON.stringify({
  status: call.status,
  taskCompleted: call.taskCompleted,
  structuredResult: call.structuredResult,
  evidence: call.evidence,
}, null, 2));
