#!/ebrmain/bin/run_script -clear_screen -bitmap=ci_autosync_cloud

APP_DIR="/mnt/ext1/applications/bookshift"
CONFIGFILE="${APP_DIR}/config.yaml"

network_up()
{
	/ebrmain/bin/netagent status > /tmp/netagent_status_wb
	read_cfg_file /tmp/netagent_status_wb NETAGENT_
	if [ "$NETAGENT_nagtpid" -gt 0 ]; then
		:
		#network enabled
	else
		/ebrmain/bin/dialog 5 "" @NeedInternet @Cancel @TurnOnWiFi
		if ! [ $? -eq 2 ]; then
			exit 1
		fi

		/ebrmain/bin/netagent net on
	fi
	/ebrmain/bin/netagent connect
}

if [ ! -f "$CONFIGFILE" ]; then
  /ebrmain/bin/dialog 5 "" "Bookshift configuration file not found: $CONFIGFILE." @OK
  exit 1
fi

# Connect to the net first if necessary.
ifconfig eth0 > /dev/null 2>&1
if [ $? -ne 0 ]; then
	touch /tmp/bookshift-wifi
	network_up
fi

"$APP_DIR"/bookshift run --config-file "$CONFIGFILE"

# Turns wifi off, if it was enabled by this script
if [ -f /tmp/bookshift-wifi ]; then *
  rm -f /tmp/bookshift-wifi
  /ebrmain/bin/netagent net off
fi

/ebrmain/cramfs/bin/scanner.app & sleep 3; kill "$!"

exit 0
