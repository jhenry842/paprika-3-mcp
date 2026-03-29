---
name: setup-aisles
description: Use this skill when the user wants to assign or fix aisles on their grocery list or pantry — "set up aisles", "assign aisles", "fix the aisle assignments", "items are in the wrong aisle", "aisles are missing". Use for both grocery list and pantry aisle setup.
---

# Setup Aisles

Bulk-assign aisles to grocery list or pantry items using the configured aisle map. Run this after adding many new items, or when items show up in the wrong aisle (or no aisle) during shopping.

## When to Run

- **After a large grocery list generation** — new items may not have aisles yet
- **After adding many pantry items** — same reason
- **When the grocery app is grouping items incorrectly** — items landed in the wrong aisle or "Unknown"
- **After the aisle map is updated** — re-run to pick up new mappings

Do not run this as routine maintenance. It's a bulk operation for when aisles need fixing.

## Step 1: Determine the Target

Infer from context, or ask if unclear:

- **Grocery list** → `setup_aisles(target="grocery")`
- **Pantry** → `setup_aisles(target="pantry")`
- **Both** → `setup_aisles(target="both")`

## Step 2: Run the Tool

Call `setup_aisles` with the appropriate `target` and `dry_run=false`. The tool returns:
- Items updated with old → new aisle
- Items with **no aisle match** — ingredients not in the aisle map

## Step 3: Handle Unknowns

If any items came back with no aisle match, present them to the user:

> These items didn't match anything in the aisle map:
> - Item A
> - Item B
>
> Want me to assign them manually? I can set aisles one at a time using `update_grocery_item_aisle`.

For each unknown item, suggest a reasonable aisle based on the ingredient type. Ask the user to confirm or correct before setting.

After assigning unknowns, note that the aisle map itself can't be updated through these tools — if the same ingredient keeps coming up as unknown, it should be added to the aisle map JSON file directly.

## Step 4: Summary

Report:
- Items updated (count)
- Items skipped — already had an aisle (count)
- Unknowns resolved manually (list)
- Unknowns still unassigned, if any

---

## Notes

- `setup_aisles` is **idempotent** — running it twice won't break anything, just skips already-assigned items.
- Manual aisle assignment uses `update_grocery_item_aisle` — it accepts multiple item UIDs at once, so batch unknowns where possible.
