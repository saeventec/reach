# Tailwind CSS

## What it is

**CSS** is the language that styles web pages: colors, sizes, spacing, layout. **Tailwind** is a CSS library of small **utility classes**, each doing one thing. You style an element by listing classes on it, instead of writing CSS rules in a separate file.

```tsx
<h1 className="text-4xl font-bold text-slate-900">Reach</h1>
```

- `text-4xl`: large text
- `font-bold`: bold
- `text-slate-900`: a very dark slate-gray color

(In React, the HTML `class` attribute is written `className`.)

## Why we use it

The owner's choice. It is very popular, the style sits right next to the markup it changes, and only the classes you use end up in the final CSS. See [ADR 0013](../decisions/0013-use-tailwind-for-styling.md).

## How we set it up

Following the official guide for Vite (Tailwind v4):

1. `npm install tailwindcss @tailwindcss/vite`
2. Add the plugin in `apps/web/vite.config.ts`: `plugins: [react(), tailwindcss()]`
3. One line in `apps/web/src/index.css`: `@import "tailwindcss";`
4. Import that CSS file once, in `src/main.tsx`: `import './index.css'`

No other config file is needed. The plugin scans our code for class names and creates CSS only for those. Our CSS file is about 5 KB.

## Common classes

| Class | Means |
| --- | --- |
| `flex`, `items-center`, `justify-center` | Lay out children in a row, centered both ways |
| `min-h-screen` | At least as tall as the browser window |
| `p-4`, `m-4`, `gap-4` | Padding, margin, gap between children (4 = 1rem, about 16 pixels) |
| `bg-slate-50`, `text-slate-900` | Background and text color (50 is lightest, 950 darkest) |
| `rounded-lg`, `shadow` | Rounded corners, a soft shadow |
| `md:text-xl` | Only on screens at least "medium" wide |

The VS Code extension "Tailwind CSS IntelliSense" (recommended in `.vscode/extensions.json`) autocompletes class names and shows the CSS behind each one.

## Alternatives

- **CSS Modules:** plain CSS files, scoped to one component. Built into Vite.
- **Plain CSS:** one global file, simplest, but styles can clash as the app grows.
- **Component libraries (MUI, shadcn/ui):** ready-made buttons and forms. Some, like shadcn/ui, are built on Tailwind.

## Learn more

- https://tailwindcss.com/docs/styling-with-utility-classes
- https://developer.mozilla.org/en-US/docs/Learn_web_development/Core/Styling_basics
