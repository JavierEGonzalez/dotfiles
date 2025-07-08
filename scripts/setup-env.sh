#!/bin/bash

# --- Function to display the override menu ---
show_override_menu() {
    clear
    echo "--- Configuration Overrides ---"
    echo "Select a setting to change. Current values are shown below."
    echo ""
    echo "--- Current Settings ---"
    echo "1. LOCAL_ENV:                 $LOCAL_ENV"
    echo "2. CF2_ENABLED:               $CF2_ENABLED"
    echo "3. DISABLE_SSR:               $DISABLE_SSR"
    echo "4. GRAPHQL_HOST_OVERRIDE:     ${GRAPHQL_HOST_OVERRIDE:-(not set)}"
    echo "5. PRISM_ENV_OVERRIDE:        ${PRISM_ENV_OVERRIDE:-(not set)}"
    echo "--------------------------"
    echo "6. Save and Finish"
    echo ""
    read -p "Enter number to modify [6]: " override_choice
}

# --- Main Script ---

# Determine opco_choice from argument or prompt
if [ -n "$1" ]; then
    opco_arg=$(echo "$1" | tr '[:lower:]' '[:upper:]')
    case $opco_arg in
        STSH) opco_choice=1 ;;
        GNTL) opco_choice=2 ;;
        GNTC) opco_choice=3 ;;
        MRTN) opco_choice=4 ;;
        FDLN) opco_choice=5 ;;
        HNFD) opco_choice=6 ;;
        *)
            echo "Invalid OPCO provided: $1"
            echo "Valid options are: STSH, GNTL, GNTC, MRTN, FDLN, HNFD"
            exit 1
            ;;
    esac
else
    # No argument, run interactively
    echo "Please select an OPCO:"
    echo "1) Stop & Shop (STSH)"
    echo "2) Giant Food (GNTL)"
    echo "3) Giant Food Stores (GNTC)"
    echo "4) Martin's Food (MRTN)"
    echo "5) Food Lion (FDLN)"
    echo "6) Hannaford (HNFD)"
    read -p "Enter number: " opco_choice
fi

override_defaults=false
if [ "$1" = "--override-defaults" ] || [ "$1" = "-d" ]; then
    override_defaults=true
    shift
elif [ "$2" = "--override-defaults" ] || [ "$2" = "-d" ]; then
    override_defaults=true
    shift
fi

# Start with a clean .env file
> .env

# OPCO specific configuration
case $opco_choice in
    1)
        echo "Setting up for Stop & Shop (STSH)"
        cat >> .env << EOL
# stop and shop
OPCO=STSH
NUXT_BR_BRAND_CHANNEL=stopandshop
TESTING_HOST=https://nuxt.alpha3.stopandshop.com:3000/
EOL
        ;;
    2)
        echo "Setting up for Giant Food (GNTL)"
        cat >> .env << EOL
# giant food
OPCO=GNTL
NUXT_BR_BRAND_CHANNEL=giant
TESTING_HOST=https://nuxt.alpha3.giantfood.com:3000/
EOL
        ;;
    3)
        echo "Setting up for Giant Food Stores (GNTC)"
        cat >> .env << EOL
# giant food stores
OPCO=GNTC
NUXT_BR_BRAND_CHANNEL=giantfoodstores
TESTING_HOST=https://nuxt.alpha3.giantfoodstores.com:3000/
EOL
        ;;
    4)
        echo "Setting up for Martin's Food (MRTN)"
        cat >> .env << EOL
# Martins food
OPCO=MRTN
NUXT_BR_BRAND_CHANNEL=martins
TESTING_HOST=https://nuxt.alpha3.martinsfoods.com:3000/
EOL
        ;;
    5)
        echo "Setting up for Food Lion (FDLN)"
        cat >> .env << EOL
# Food lion
OPCO=FDLN
NUXT_BR_BRAND_CHANNEL=foodlion
TESTING_HOST=https://nuxt.alpha3.foodlion.com:3000/
EOL
        ;;
    6)
        echo "Setting up for Hannaford (HNFD)"
        cat >> .env << EOL
# hannaford
OPCO=HNFD
NUXT_BR_BRAND_CHANNEL=hannaford
TESTING_HOST=https://nuxt.alpha3.hannaford.com:3000/
EOL
        ;;
    *)
        echo "Invalid choice. Exiting."
        rm .env
        exit 1
        ;;

esac

# --- SSH Certs ---
if [ -f ".env.certs" ]; then
    echo "Found .env.certs, adding to .env file."
    echo "" >> .env
    cat .env.certs >> .env
    # source the certs to make them available in the script
    . .env.certs
else
    echo "--- SSH Configuration ---"
    read -p "Enter path for SSH_KEY: " SSH_KEY
    read -p "Enter path for SSH_CERT: " SSH_CERT

    read -p "Save these paths for future use? (y/N): " save_certs_choice
    if [[ "$save_certs_choice" =~ ^[Yy]$ ]]; then
        echo "Saving SSH paths to .env.certs"
        echo "SSH_KEY=$SSH_KEY" > .env.certs
        echo "SSH_CERT=$SSH_CERT" >> .env.certs
    fi
    echo "\nSSH_KEY=$SSH_KEY" >> .env
    echo "SSH_CERT=$SSH_CERT" >> .env
fi

# --- Set Default Values ---
LOCAL_ENV="local"
CF2_ENABLED="false"
DISABLE_SSR="false"
GRAPHQL_HOST_OVERRIDE=""
PRISM_ENV_OVERRIDE=""

# --- Override Loop ---
if [ "$override_defaults" = "true" ]; then
    while true; do
        show_override_menu
        case $override_choice in
            1)
                echo "Select an environment:"
                echo "1) local (default), 2) dev, 3) qa, 4) prod"
                read -p "Enter number [1]: " env_choice
                case $env_choice in
                    2) LOCAL_ENV="dev" ;;
                    3) LOCAL_ENV="qa" ;;
                    4) LOCAL_ENV="prod" ;;
                    *) LOCAL_ENV="local" ;;
                esac
                ;;
            2)
                read -p "Enable CF2 environments? (y/N): " cf2_choice
                [[ "$cf2_choice" =~ ^[Yy]$ ]] && CF2_ENABLED="true" || CF2_ENABLED="false"
                ;;
            3)
                read -p "Disable Server-Side Rendering (SSR)? (y/N): " ssr_choice
                [[ "$ssr_choice" =~ ^[Yy]$ ]] && DISABLE_SSR="true" || DISABLE_SSR="false"
                ;;
            4)
                read -p "Enter GraphQL host override (or press Enter to unset): " GRAPHQL_HOST_OVERRIDE
                ;;
            5)
                read -p "Enter Prism environment override (or press Enter to unset): " PRISM_ENV_OVERRIDE
                ;;
            6 | "")
                break
                ;;
            *)
                echo "Invalid option. Please try again."
                sleep 1
                ;;
        esac
    done
fi

# Advanced Overrides
echo "" >> .env
echo "# leave blank to compute based on environment" >> .env
if [ -n "$GRAPHQL_HOST_OVERRIDE" ]; then
    echo "GRAPHQL_HOST_OVERRIDE="$GRAPHQL_HOST_OVERRIDE"" >> .env
fi
if [ -n "$PRISM_ENV_OVERRIDE" ]; then
    echo "PRISM_ENV_OVERRIDE="$PRISM_ENV_OVERRIDE"" >> .env
fi

# Environment
echo "" >> .env
echo "# "local" environment will set some feature flags to true." >> .env
echo "LOCAL_ENV=$LOCAL_ENV" >> .env
echo "" >> .env

# CF2
echo "CF2_ENABLED=$CF2_ENABLED" >> .env
echo "" >> .env

# SSR
if [ "$DISABLE_SSR" = "true" ]; then
    echo "DISABLE_SSR=true" >> .env
else
    echo "# DISABLE_SSR=true" >> .env
fi
echo "" >> .env

# Append static config
cat >> .env << EOL
USER_AGENT="Prism Web Client/N;u+x.t 1.0"
NUXT_GOOGLE_MAPS_API_KEY=test
OPTIMIZELY_SDK_KEY_DEV=XuCjuTaJydn8fg7ubKPuP
OPTIMIZELY_SDK_KEY_PROD=KzYgRRrDfuLjwqk1fjSaG
NUXT_BR_ENDPOINT=https://peapod.bloomreach.io
NUXT_BR_PREVIEW_TOKEN=
ONE_TRUST_ID=''
EOL

echo "" >> .env
echo ""
echo ".env file created successfully."
