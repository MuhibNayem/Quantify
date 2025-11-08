#!/bin/bash

# Navigate to the client directory
cd /Users/a.k.mmuhibullahnayem/Developer/Quantify/client

# Function to add a shadcn-svelte component
add_component() {
  COMPONENT=$1
  echo "Adding shadcn-svelte component: $COMPONENT"
  # Run the command and automatically input 'Yes'
  echo "Yes" | pnpm dlx shadcn-svelte@latest add "$COMPONENT"
  if [ $? -ne 0 ]; then
    echo "Error adding $COMPONENT. Please check the output above."
    exit 1
  fi
}

# Add dialog component
add_component dialog

# Add table component
add_component table

# Add form component
add_component form

# Add dropdown-menu component
add_component dropdown-menu

# Add sheet component
add_component sheet

# Add tabs component
add_component tabs

# Add alert component
add_component alert

# Add badge component
add_component badge

# Add pagination component
add_component pagination

# Add calendar component
add_component calendar

# Add checkbox component
add_component checkbox

# Add switch component
add_component switch

# Add textarea component
add_component textarea

# Add avatar component
add_component avatar

# Add separator component
add_component separator

# Add popover component
add_component popover

# Add Card component
add_component card

# Add Input component
add_component input

# Add Label component
add_component label

# Add Button component
add_component button

# Add ContextMenu component
add_component context-menu

# Add Commad component
add_component command

# Add Pagination component
add_component pagination

# Add Tooltip component
add_component tooltip

# Add Skeleton component
add_component skeleton

# Add Badge component
add_component badge

# Add NavigationMenu component
add_component navigation-menu

# Add Select component
add_component select

echo "All specified shadcn-svelte components have been added (or attempted to be added)."
echo "Please review the output for any errors."
