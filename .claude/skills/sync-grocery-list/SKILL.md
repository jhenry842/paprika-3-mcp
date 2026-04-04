---
name: sync-grocery-list
description: Use this skill when the user is done shopping and wants to sync their grocery list to the pantry — "I'm done shopping", "sync the grocery list", "mark everything as bought", "update the pantry from my shopping", or any variation. Always use this skill for post-shopping pantry sync, not the raw tool.
---

# Sync Grocery List to Pantry

Post-shopping workflow: confirm what was purchased, sync all checked items to the pantry, uncheck staples so they stay ready for next week, delete non-staples, and prompt for next steps.

## Step 0: Confirm This Is a Mid-Cycle Top-Up

Before doing anything, ask:

> "Was this a mid-week top-up run, or are you wrapping up the cooking cycle for the week?"

- **Mid-week top-up** (ran out of something, quick extra trip) → continue with this skill.
- **Wrapping up the cycle** (done cooking for the week, ready to plan next week) → stop and say: "Use `close-cycle` instead — it handles pantry depletion from cooked meals and advances the sync date. Just say 'close the cycle'."

Do not proceed until the user confirms this is a top-up.

## Step 1: Preview What Will Be Synced

Call `get_grocery_list`. Separate items into:
- **Checked (purchased=true)** — these will be processed
- **Unchecked** — these remain untouched

If there are **no checked items**, stop and tell the user — nothing to sync. Suggest they check off items in the Paprika app first.

Present a brief summary before doing anything:

> You have **N checked items** ready to sync to the pantry:
> - Item A (staple)
> - Item B
> - ...
>
> **M items** will remain on the grocery list (unchecked).
>
> Proceed?

Wait for confirmation before proceeding.

## Step 2: Identify Staples

Call `get_household_rules`. Look for rules with `type: "staple"` — each has `params.ingredient` containing a canonical lowercase ingredient name (e.g. `"apples"`, `"bananas"`).

For each checked grocery item, check if its `ingredient` field (lowercase) matches any staple rule. Build two lists:
- **Staples** — match a staple rule; will be unchecked after pantry sync
- **Non-staples** — will be deleted after pantry sync

## Step 3: Sync All Checked Items to Pantry

Call `get_pantry` once. For every checked item (both staples and non-staples):

1. Resolve the ingredient name: use `ingredient` if non-empty, otherwise fall back to `name`.
2. Check if the ingredient exists in the pantry (case-insensitive match on `ingredient` field).
   - **Exists:** call `update_pantry_item` with `in_stock: true` and the grocery item's quantity (if non-empty).
   - **Doesn't exist:** call `add_pantry_item` with `ingredient`, `quantity`, and `in_stock: true`.

## Step 4: Uncheck Staples

If any staples were identified, call `uncheck_grocery_items` with their UIDs. This sets `purchased=false` so they remain on the list, unchecked, for the next shopping trip.

## Step 5: Delete Non-Staples

Call `delete_grocery_items` with the UIDs of all non-staple checked items.

## Step 5b: Record Purchase History

Silently record what was purchased this trip so future grocery generation can surface recurring patterns.

1. From the checked items processed in Steps 3–5, collect the `ingredient` field (lowercase) for each item — both staples and non-staples. Deduplicate.
2. Call `get_household_rules`. Find the rule with `id: "grocery-history"`. If it exists, extract `params.cycles`; if not, start with an empty array.
3. Append a new record: `{ "cycle_end": "<today YYYY-MM-DD>", "purchased": [<ingredient list>] }`.
4. Trim the array to the most recent 5 entries.
5. Call `set_household_rule` with `id: "grocery-history"`, `type: "history"`, `description: "Purchase history across recent grocery cycles — used to suggest recurring items during grocery generation"`, and the updated `params.cycles`.

Do this silently — no user-facing output.

## Step 6: Summary

Report what happened concisely:

- **Added to pantry** — new items
- **Updated in pantry** — existing items marked back in-stock
- **Kept on list (staples)** — unchecked and ready for next week
- **Removed from grocery list** — count cleared
- **Errors**, if any

## Step 7: Prompt Next Steps

Ask: "Want me to generate the grocery list for next week now, or wait until after you've planned meals?"

If yes, run the `generate-grocery-list` skill inline.

If no: "When you're ready, say 'plan the week' or 'generate grocery list'."

---

## Notes

- **This skill does NOT advance `last_sync_date` and does NOT deplete the pantry.** It is for mid-cycle top-up shops only. For end-of-cycle close, use the `close-cycle` skill instead — that runs depletion + restock + sync date advancement as one atomic operation.
- **Never sync without showing the checked item list first.** The user may have accidentally checked something.
- **Staple matching is by `ingredient` field, not display name.** "Apples (3 lbs)" has ingredient `"apples"` — match on that.
- **If `get_household_rules` returns no staple rules**, skip Step 2 and treat everything as non-staple.
- **If a tool call fails**, report it clearly and continue processing the remaining items. Do not abort the whole sync on a single failure.
- **To add a new staple**, the user can say "add apples as a staple" — call `set_household_rule` with `type: "staple"`, `id: "staple-apples"`, `description: "Apples are a weekly staple — keep on list after shopping"`, `params: {"ingredient": "apples"}`.
