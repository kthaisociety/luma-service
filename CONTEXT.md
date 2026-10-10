# Vocabulary

Words as this project uses them. Luma's own names are in parentheses where they differ.

- **Luma**: the event and community platform KTHAIS runs its calendar on (`luma.com`).
- **Calendar**: the KTHAIS calendar on Luma. The API key belongs to it, so the service reads exactly one.
- **Event**: an event listed on the calendar. `platform` is `luma` for Luma events; others are external
  events added to the calendar.
- **Tier** (membership tier): a membership level on the calendar that people apply to join, with its own
  registration questions.
- **Community tier**: the tier the growth team analyses, set by `LUMA_COMMUNITY_TIER_ID`.
- **Contact**: anyone known to the calendar (guests, subscribers, members).
- **Member**: a contact whose membership in a tier is `approved`. The service only serves approved members.
- **Registration question**: a question asked when joining a tier, e.g. graduation year or areas of interest.
- **Answer** (registration answer): a member's answer to one question, always a list of strings here
  ([ADR 0001](docs/adr/0001-answers-as-string-lists.md)).
- **Caller**: another KTHAIS service using this one, e.g. growth-portal. Authenticates with the
  **service secret**.
