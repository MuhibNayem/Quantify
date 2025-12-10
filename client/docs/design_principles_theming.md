# Design Principles & Theming Rules

## Apple-Liquid-Glass Aesthetic (Advanced 3D)

The goal is to move beyond "flat glass" to a **physical, 3D glass presence**. Elements should feel like they have weight, thickness, and interact with light.

### 1. Physical Presence & Lighting
- **3D Borders**: Borders must simulate a directional light source (top-left).
    - **Top/Left**: Lighter/Opaque (e.g., `rgba(255, 255, 255, 0.65)`) to catch the light.
    - **Bottom/Right**: Darker/Transparent (e.g., `rgba(255, 255, 255, 0.15)`) to cast a shadow.
    - *Constraint*: **Never use inline Tailwind border classes** (e.g., `border`, `border-white/40`) on liquid elements. They override the custom 3D CSS borders.
- **Elevation**: Use multi-layer shadows. A soft drop shadow for elevation + an inner white inset shadow to depict the glass edge thickness.

### 2. Soothing Transparent Gradients (No Flat Opacity)
- **Backgrounds**: Avoid flat `bg-white/XX`. Use soothing transparent gradients to create flow and depth.
    - **Standard Formula**: `bg-gradient-to-br from-white/40 via-white/20 to-white/5`
    - This allows the background glow/blur to shine through variably, creating a more organic "liquid" feel.

### 3. Organic Interactions
- **Hover Lift**: Interactive elements (`.liquid-hoverable`) should physically "lift" (translateY) and catch more light (border brightens) on hover.
- **Inner Glow**: Transitions should prioritize `box-shadow` and `backdrop-filter` changes over simple color swaps.

## Components & Classes

- **.liquid-panel**: The container for main sections.
    - *Must Use*: The transparent gradient formula.
    - *Must Avoid*: Inline borders.
- **.liquid-hoverable**: For lists, cards, and interactive rows.
    - Includes built-in 3D hover physics.
    - Use for anything that is clickable or needs to stand out.
- **.glass-toggle-active**: Use an overlay/underlay div for active states, fading it in/out. Do not animate the container class itself to avoid layout shifts.

## UX Copy
- **Value First**: "Time Intelligence" instead of "Clock In".
- **Invisible Tech**: Never mention "CSS", "Glass", or implementation details in user-facing text.

## Critical Implementation Rules
1. **No Card-on-Card Stacking**: Don't put opaque white cards inside glass panels.
2. **Conflict Avoidance**: If using `.liquid-panel` or `.liquid-hoverable`, **REMOVE** any `border-*` or `shadow-*` utility classes from the HTML. Let the global CSS handle the physics.
