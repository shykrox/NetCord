# ComfyUI Worker Plan

Current state:

- `/ask` and `/draw` chat commands create `ai_jobs`.
- `POST /ai/ask` and `POST /ai/draw` create jobs explicitly.
- `GET /ai/jobs` and `GET /ai/jobs/{id}` expose job state.
- WebSocket job events are available.
- The Qt client displays AI job cards.

Scaffold-only part:

- No background worker consumes `ai_jobs` yet.
- No generated image is uploaded to MinIO automatically yet.

Recommended worker flow:

1. Poll or listen for queued `ai_jobs`.
2. Mark job `running`, broadcast `ai.job.updated`.
3. For `/draw`, submit workflow to `NETCORD_COMFYUI_URL`.
4. Track progress and broadcast updates.
5. Download generated output.
6. Upload output to MinIO using the existing attachment pipeline.
7. Create a bot-authored message with the generated attachment.
8. Mark job `completed` and broadcast `job.completed`.

Keep ComfyUI internal URLs and credentials server-side. Do not expose them to the Qt client.
