# XP Themes

This directory contains theme files for the XP file explorer. Themes are defined in JSON format and loaded dynamically at startup.

## Theme Structure

Each theme file must follow this JSON structure:

```json
{
  "name": "ThemeName",
  "colors": {
    "text": "white",
    "background": "black",
    "highlight": "magenta",
    "highlight_text": "white",
    "footer": "cyan",
    "footer_bg": "black",
    "address_bar": "magenta",
    "address_bar_bg": "black",
    "separator": "magenta",
    "dim": "white",
    "filter": "white",
    "filter_bg": "magenta",
    "dir": "cyan"
  }
}
```

## Available Colors

- `black`
- `red`
- `green`
- `yellow`
- `blue`
- `magenta`
- `cyan`
- `white`
- `default` (terminal default)

## Color Validation

The theme system automatically validates that:
- Text and background colors are different
- Footer text and footer background are different
- Address bar text and address bar background are different
- Filter text and filter background are different
- Highlight text and highlight background are different

If any pair has the same color, the system will automatically adjust the text color to ensure readability.

## Creating Custom Themes

1. Create a new `.json` file in this directory
2. Follow the structure above
3. Choose a unique name
4. Ensure text/background color pairs are different
5. Restart XP or press 'O' to see your new theme

## Included Themes

- **Nightfall** - Dark theme with magenta accents
- **Forest** - Green nature theme
- **Ocean** - Blue water theme
- **Sunset** - Warm red and yellow theme
- **Lavender** - Purple and cyan theme
- **Autumn** - Yellow and red fall colors

## Tips

- Use contrasting colors for better readability
- Test your theme in different lighting conditions
- Nature-inspired themes work well with the file explorer aesthetic