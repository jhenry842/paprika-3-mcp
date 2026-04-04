---
name: pantry-review
description: Use this skill when the user asks for a pantry check, "what's in my pantry?", "pantry health check", "what do I have?", "what proteins do I have?", "what's low?", or wants a mid-cycle review of pantry status without closing the cycle.
---

# Pantry Review

On-demand mid-cycle pantry health check. Surfaces missing proteins, low-stock staples, and anything worth flagging before the next shop. Does NOT modify the pantry or advance last_sync_date.

## Step 1: Get Pantry and Rules

Call `get_pantry` and `get_household_rules` in parallel.

From the rules, collect all `staple` entries — these are ingredients that should always be on hand.

---

## Step 2: Check Proteins

Scan the pantry for protein items. Proteins are: beef (ground, stew, roast, brisket, steak, short ribs), chicken (breast, thigh, drumstick, whole, ground), pork (chops, tenderloin, ground, sausage, bacon, ham), fish and seafood (salmon, cod, tilapia, shrimp, tuna, halibut), lamb, turkey, and any household substitutions (e.g. venison for ground beef — check substitution rules).

Bucket each protein as:
- **In stock** — `in_stock: true` with a non-empty quantity
- **Out of stock** — `in_stock: false`
- **Unknown quantity** — `in_stock: true` but quantity is blank or unclear

---

## Step 3: Check Staples

For each staple rule, find the matching pantry item (case-insensitive, substring match is fine).

Bucket each staple as:
- **In stock**
- **Out of stock or missing from pantry entirely**

---

## Step 4: Flag Notable Items

Look across the full pantry for anything worth flagging:
- Items that are `in_stock: false` and haven't been mentioned in Steps 2–3
- Items with very low quantities (e.g., "1 tbsp", "1 clove", "a pinch") that will run out soon
- Items that appear to be duplicates (same ingredient, different entries)

Keep this list short — only genuinely notable items. Don't flag every out-of-stock item in an exhaustive dump.

---

## Step 5: Present the Report

Report in three sections. Omit any section that has nothing to show.

### Proteins
One line per protein group (beef, chicken, pork, etc.). Show in-stock items with quantity; call out out-of-stock ones clearly.

Example:
> - **Chicken**: breast (3 lbs) ✓, thighs — out of stock
> - **Beef/Venison**: ground venison (2 lbs) ✓, venison roast — out of stock
> - **Pork**: none in stock

### Staples
Flag any staple that's missing or out of stock. If all staples are covered, say so in one line.

Example:
> All staples in stock.

Or:
> - **Berries** — out of stock
> - **Carrots** — not in pantry

### Worth Noting
Any low-quantity or flagged items from Step 4. Keep to 3–5 items max.

---

## Step 6: Follow-Up

After the report, offer one next step based on what was found:

- **If proteins are missing:** "Want me to add those to the grocery list?"
- **If staples are out:** "Want me to add the missing staples to the grocery list?"
- **If pantry looks healthy:** "Pantry looks good — want to run plan-the-week or check what you can make tonight?"

Only offer the most relevant follow-up. Don't list all three options.

---

## Notes

- **Read-only.** This skill never modifies the pantry, grocery list, or any rules.
- **Mid-cycle only.** For end-of-cycle depletion + restock, use `close-cycle`.
- **Substitutions matter.** If there's a substitution rule (e.g. venison for ground beef), apply it when evaluating proteins — "ground beef out of stock" is not a gap if venison is in stock.
- **Don't be exhaustive.** The goal is a quick health snapshot, not a full inventory printout. Surface what matters, skip the noise.
