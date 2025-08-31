# Font Customization

This document describes how to configure custom fonts in pgweb for company branding and improved user experience.

## Overview

pgweb supports customizable fonts through environment variables, allowing each deployment to use company-specific typography. The system supports:

- Any Google Fonts font family
- Custom font sizes
- Automatic font loading from Google Fonts API
- Font persistence across user sessions
- Application to both interface text and code editor

## Environment Variables

### `PGWEB_FONT_FAMILY`

**Description:** CSS font family to use throughout the interface  
**Format:** CSS font-family string  
**Default:** System fonts (`-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif`)

**Examples:**

```bash
export PGWEB_FONT_FAMILY="Space Grotesk, sans-serif"
export PGWEB_FONT_FAMILY="Inter, sans-serif"
export PGWEB_FONT_FAMILY="Roboto, sans-serif"
export PGWEB_FONT_FAMILY="'Custom Font Name', Arial, sans-serif"
```

### `PGWEB_FONT_SIZE`

**Description:** CSS font size for the interface  
**Format:** CSS size value with units  
**Default:** `14px`

**Examples:**

```bash
export PGWEB_FONT_SIZE="14px"
export PGWEB_FONT_SIZE="15px"
export PGWEB_FONT_SIZE="16px"
export PGWEB_FONT_SIZE="1rem"
```

### `PGWEB_GOOGLE_FONTS`

**Description:** Google Fonts to preload for the interface  
**Format:** Comma-separated list of `FontName:weights`  
**Default:** None (only loads if specified)

**Examples:**

```bash
# Single font with multiple weights
export PGWEB_GOOGLE_FONTS="Space Grotesk:300,400,500,600,700"

# Multiple fonts
export PGWEB_GOOGLE_FONTS="Inter:300,400,500,600,700,Roboto:300,400,500,700"

# Font with specific weights
export PGWEB_GOOGLE_FONTS="Poppins:400,600,700"
```

## Configuration Methods

### 1. Docker Entrypoint Script

Add font configuration to your Docker entrypoint script:

```bash
#!/bin/sh
set -e

# Font Configuration - Company branding
export PGWEB_FONT_FAMILY="Space Grotesk, sans-serif"
export PGWEB_FONT_SIZE="15px"
export PGWEB_GOOGLE_FONTS="Space Grotesk:300,400,500,600,700"

# Your existing pgweb configuration...
exec pgweb --bind=0.0.0.0 --listen=8081 [other options]
```

### 2. Docker Compose

Configure fonts in your `docker-compose.yml`:

```yaml
services:
  pgweb:
    image: your-pgweb-image
    environment:
      PGWEB_FONT_FAMILY: "Space Grotesk, sans-serif"
      PGWEB_FONT_SIZE: "15px"
      PGWEB_GOOGLE_FONTS: "Space Grotesk:300,400,500,600,700"
      # Your other environment variables...
```

### 3. Environment File

Create a `.env` file:

```bash
# Font Configuration
PGWEB_FONT_FAMILY=Space Grotesk, sans-serif
PGWEB_FONT_SIZE=15px
PGWEB_GOOGLE_FONTS=Space Grotesk:300,400,500,600,700
```

### 4. Command Line Arguments

You can also use command line flags:

```bash
pgweb --font-family="Space Grotesk, sans-serif" --font-size="15px" --google-fonts="Space Grotesk:300,400,500,600,700"
```

## Popular Font Examples

### Tech Companies

```bash
# Modern, clean fonts
export PGWEB_FONT_FAMILY="Inter, sans-serif"
export PGWEB_GOOGLE_FONTS="Inter:300,400,500,600,700"

export PGWEB_FONT_FAMILY="Poppins, sans-serif"
export PGWEB_GOOGLE_FONTS="Poppins:300,400,500,600,700"
```

### Financial/Corporate

```bash
# Professional, readable fonts
export PGWEB_FONT_FAMILY="Roboto, sans-serif"
export PGWEB_GOOGLE_FONTS="Roboto:300,400,500,700"

export PGWEB_FONT_FAMILY="Open Sans, sans-serif"
export PGWEB_GOOGLE_FONTS="Open Sans:300,400,600,700"
```

### Creative/Design

```bash
# Distinctive, branded fonts
export PGWEB_FONT_FAMILY="Space Grotesk, sans-serif"
export PGWEB_GOOGLE_FONTS="Space Grotesk:300,400,500,600,700"

export PGWEB_FONT_FAMILY="Nunito, sans-serif"
export PGWEB_GOOGLE_FONTS="Nunito:300,400,600,700,800"
```

## Technical Details

### Font Loading Process

1. **Environment Variables** are read by the Go backend during startup
2. **Configuration API** (`/api/config`) exposes font settings to the frontend
3. **JavaScript** fetches configuration and dynamically loads Google Fonts
4. **CSS Custom Properties** apply fonts throughout the interface
5. **Ace Editor** receives font updates for consistent code styling

### Browser Support

The font system uses modern web standards:

- CSS Custom Properties (CSS Variables)
- Google Fonts API v2
- Dynamic font loading via JavaScript
- localStorage for persistence

### Performance Considerations

- Fonts are loaded asynchronously to prevent blocking
- `font-display: swap` ensures immediate text rendering
- Google Fonts are cached by the browser
- Only specified font weights are loaded to minimize bandwidth

### Fallback Strategy

The system includes robust fallbacks:

1. **Configured font** (from environment)
2. **System fonts** (if Google Font fails to load)
3. **Generic sans-serif** (ultimate fallback)

## Troubleshooting

### Font Not Loading

- Check that the font name exactly matches Google Fonts spelling
- Verify environment variables are properly set in your deployment
- Check browser developer tools for font loading errors
- Ensure internet connectivity for Google Fonts API

### Font Not Applying to Code Editor

- The Ace editor font is automatically updated when the main font changes
- If issues persist, check browser console for JavaScript errors

### Configuration Not Taking Effect

- Restart the pgweb service after changing environment variables
- Verify the `/api/config` endpoint returns your font configuration
- Clear browser cache if fonts appear to be cached incorrectly

## API Reference

### GET /api/config

Returns current configuration including font settings:

```json
{
  "fonts": {
    "family": "Space Grotesk, sans-serif",
    "size": "15px",
    "google_fonts": "Space Grotesk:300,400,500,600,700"
  },
  "parameter_patterns": {
    "custom": ["Client", "Instance", ...]
  }
}
```
