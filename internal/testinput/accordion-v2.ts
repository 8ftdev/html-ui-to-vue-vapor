/**
 * Accordion
 * Native disclosure with an accessible summary.
 * Behavior: native. See `html-ui --docs accordion`.
 * Source conversion contract v2; native factory initializes once.
 */
export const contractVersion = 2 as const;

export interface AccordionProps {
  /** Nonempty name groups details exclusively within the same tree. */
  name?: string;
  /** Whether the disclosure is initially open. */
  open?: boolean;
}

export const defaults = {
  open: false,
} as const satisfies Partial<AccordionProps>;

export interface AccordionSlots<Content = HTMLElement> {
  /** Accessible disclosure label; avoid nested interactive controls. */
  summary: () => Content;
  /** Disclosure body. */
  content?: (scope: { open: boolean }) => Content;
}

export interface AccordionEvents {
  /** Native notification; adapters own application state updates. */
  toggle: ToggleEvent;
}

/** Native event targets and DOM properties to read after an event. No listeners are installed. */
export const nativeEvents = {
  toggle: { target: "root", state: { open: "open", } },
} as const;

/** Portable anatomy and state sources. Preserve through framework and theme transforms. */
export const ui = {
  "component": "accordion",
  "behavior": {
    "kind": "native",
    "requirements": [
      "The summary is the first child. Its native activation opens or closes the details.",
      "Read root.open on toggle to synchronize adapter state; slot scope is an initial snapshot in the native factory.",
      "For exclusive groups use the same nonempty name in the same tree, not necessarily siblings. Do not nest members of the same named group."
    ]
  },
  "parts": {
    "root": {
      "node": "root",
      "styleRole": "disclosure",
      "state": {
        "expanded": {
          "source": {
            "node": "root",
            "attribute": "open",
            "present": true
          }
        }
      }
    },
    "trigger": {
      "node": "summary",
      "styleRole": "disclosure-trigger",
      "state": {
        "expanded": {
          "source": {
            "node": "root",
            "attribute": "open",
            "present": true
          }
        },
        "focusVisible": {
          "source": {
            "node": "summary",
            "pseudo": "focus-visible"
          }
        },
        "hover": {
          "source": {
            "node": "summary",
            "pseudo": "hover"
          }
        }
      }
    }
  }
} as const;

/** Styling operations only; adapters implement composition without changing native behavior. */
export type AccordionStyle<Style = string> = Style | { mode: "replace"; value: Style } | { mode: "omit" };

/** Owned styling targets only. Content slots are not implicitly wrapped. */
export interface AccordionClasses<Style = string> {
  root?: {
    base?: AccordionStyle<Style>;
    unstyled?: boolean;
    state?: {
      expanded?: AccordionStyle<Style>;
    };
  };
  trigger?: {
    base?: AccordionStyle<Style>;
    unstyled?: boolean;
    state?: {
      expanded?: AccordionStyle<Style>;
      focusVisible?: AccordionStyle<Style>;
      hover?: AccordionStyle<Style>;
    };
  };
}

/** Build initial native structure; converters parse this body without executing it. */
export function accordion(props: AccordionProps, slots: AccordionSlots<HTMLElement>): HTMLDetailsElement {
  const { name, open = defaults.open } = props;
  const root = document.createElement("details");
  const summary = document.createElement("summary");
  root.setAttribute("data-ui", "accordion");
  root.setAttribute("data-ui-part", "root");
  summary.setAttribute("data-ui", "accordion");
  summary.setAttribute("data-ui-part", "trigger");
  if (name !== undefined) root.setAttribute("name", name);
  root.open = open;
  summary.append(slots.summary());
  root.append(summary);
  if (slots.content !== undefined) root.append(slots.content({ open }));
  return root;
}
