---
name: setup-aisles
description: Use this skill when the user wants to assign or fix aisles on their grocery list or pantry — "set up aisles", "assign aisles", "fix the aisle assignments", "items are in the wrong aisle", "aisles are missing". Use for both grocery list and pantry aisle setup.
---

# Setup Aisles

Bulk-assign aisles to grocery list or pantry items using the Woodman's East aisle map. Run this after adding many new items, or when items show up in the wrong aisle (or no aisle) during shopping.

## When to Run

- **After a large grocery list generation** — new items may not have aisles yet
- **After adding many pantry items** — same reason
- **When the grocery app is grouping items incorrectly** — items landed in the wrong aisle or "Unknown"
- **After the aisle map is updated** — re-run to pick up new mappings

Do not run this as routine maintenance. It's a bulk operation for when aisles need fixing.

## Step 1: Determine the Target

Ask the user (or infer from context):

> Do you want to set up aisles for the **grocery list**, the **pantry**, or both?

- **Grocery list** → call `setup_woodmans_aisles`
- **Pantry** → call `setup_pantry_aisles`
- **Both** → call both, grocery list first

## Step 2: Run the Tool(s)

Call the appropriate tool(s). Each tool returns:
- Count of items updated
- Count of items skipped (already had an aisle)
- List of items with **no aisle match** — ingredients not in the Woodman's East aisle map

## Step 3: Handle Unknowns

If any items came back with no aisle match, present them to the user:

> These items didn't match anything in the aisle map:
> - Item A
> - Item B
>
> Want me to assign them manually? I can set aisles one at a time using `update_grocery_item_aisle`.

For each unknown item, suggest a reasonable aisle based on the ingredient type (e.g., "dried lentils" → Dry Goods, "kombucha" → Beverages). Ask the user to confirm or correct before setting.

After assigning unknowns, note that the aisle map itself can't be updated through these tools — if the same ingredient keeps coming up as unknown, it should be added to `aisles/woodmans_east.json` directly.

## Step 4: Summary

Report:
- Items updated (count)
- Items skipped — already had an aisle (count)
- Unknowns resolved manually (list)
- Unknowns still unassigned, if any

---

## Notes

- The aisle map is Woodman's East specific. If shopping at a different store, aisle assignments will be wrong — skip this step.
- `setup_woodmans_aisles` and `setup_pantry_aisles` are **idempotent** — running them twice won't break anything, just skips already-assigned items.
- Manual aisle assignment uses `update_grocery_item_aisle` — it accepts multiple item UIDs at once, so batch unknowns where possible.
