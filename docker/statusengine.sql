-- phpMyAdmin SQL Dump
-- version 4.5.4.1deb2ubuntu2.1
-- http://www.phpmyadmin.net
--
-- Host: localhost
-- Erstellungszeit: 07. Okt 2018 um 21:03
-- Server-Version: 5.7.23-0ubuntu0.16.04.1
-- PHP-Version: 7.0.32-0ubuntu0.16.04.1

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Datenbank: `statusengine`
--

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_dbversion`
--

CREATE TABLE `statusengine_dbversion` (
  `id` int(13) NOT NULL,
  `dbversion` varchar(255) DEFAULT '3.0.0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_hostchecks`
--

CREATE TABLE `statusengine_hostchecks` (
  `hostname` varchar(255) DEFAULT NULL,
  `state` tinyint(2) UNSIGNED DEFAULT '0',
  `is_hardstate` tinyint(1) UNSIGNED DEFAULT '0',
  `start_time` bigint(13) NOT NULL,
  `end_time` bigint(13) NOT NULL,
  `output` varchar(1024) DEFAULT NULL,
  `timeout` tinyint(3) UNSIGNED DEFAULT '0',
  `early_timeout` tinyint(1) UNSIGNED DEFAULT '0',
  `latency` float DEFAULT '0',
  `execution_time` float DEFAULT '0',
  `perfdata` varchar(1024) DEFAULT NULL,
  `command` varchar(1024) DEFAULT NULL,
  `current_check_attempt` tinyint(3) UNSIGNED DEFAULT '0',
  `max_check_attempts` tinyint(3) UNSIGNED DEFAULT '0',
  `long_output` varchar(8192) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_hoststatus`
--

CREATE TABLE `statusengine_hoststatus` (
  `hostname` varchar(255) NOT NULL,
  `status_update_time` bigint(13) NOT NULL,
  `output` varchar(1024) DEFAULT NULL,
  `long_output` varchar(1024) DEFAULT NULL,
  `perfdata` varchar(1024) DEFAULT NULL,
  `current_state` tinyint(2) UNSIGNED DEFAULT '0',
  `current_check_attempt` tinyint(3) UNSIGNED DEFAULT '0',
  `max_check_attempts` tinyint(3) UNSIGNED DEFAULT '0',
  `last_check` bigint(13) NOT NULL,
  `next_check` bigint(13) NOT NULL,
  `is_passive_check` tinyint(1) UNSIGNED DEFAULT '0',
  `last_state_change` bigint(13) NOT NULL,
  `last_hard_state_change` bigint(13) NOT NULL,
  `last_hard_state` tinyint(2) UNSIGNED DEFAULT '0',
  `is_hardstate` tinyint(1) UNSIGNED DEFAULT '0',
  `last_notification` bigint(13) NOT NULL,
  `next_notification` bigint(13) NOT NULL,
  `notifications_enabled` tinyint(1) UNSIGNED DEFAULT '0',
  `problem_has_been_acknowledged` tinyint(1) UNSIGNED DEFAULT '0',
  `acknowledgement_type` tinyint(2) UNSIGNED DEFAULT '0',
  `passive_checks_enabled` tinyint(1) UNSIGNED DEFAULT '0',
  `active_checks_enabled` tinyint(1) UNSIGNED DEFAULT '0',
  `event_handler_enabled` tinyint(1) UNSIGNED DEFAULT '0',
  `flap_detection_enabled` tinyint(1) UNSIGNED DEFAULT '0',
  `is_flapping` tinyint(1) UNSIGNED DEFAULT '0',
  `latency` float DEFAULT '0',
  `execution_time` float DEFAULT '0',
  `scheduled_downtime_depth` tinyint(2) UNSIGNED DEFAULT '0',
  `process_performance_data` tinyint(1) UNSIGNED DEFAULT '0',
  `obsess_over_host` tinyint(1) UNSIGNED DEFAULT '0',
  `normal_check_interval` int(11) UNSIGNED DEFAULT '0',
  `retry_check_interval` int(11) UNSIGNED DEFAULT '0',
  `check_timeperiod` varchar(255) DEFAULT NULL,
  `node_name` varchar(255) DEFAULT NULL,
  `last_time_up` bigint(13) NOT NULL,
  `last_time_down` bigint(13) NOT NULL,
  `last_time_unreachable` bigint(13) NOT NULL,
  `current_notification_number` int(11) UNSIGNED DEFAULT '0',
  `percent_state_change` double DEFAULT '0',
  `event_handler` varchar(255) DEFAULT NULL,
  `check_command` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Daten für Tabelle `statusengine_hoststatus`
--

INSERT INTO `statusengine_hoststatus` (`hostname`, `status_update_time`, `output`, `long_output`, `perfdata`, `current_state`, `current_check_attempt`, `max_check_attempts`, `last_check`, `next_check`, `is_passive_check`, `last_state_change`, `last_hard_state_change`, `last_hard_state`, `is_hardstate`, `last_notification`, `next_notification`, `notifications_enabled`, `problem_has_been_acknowledged`, `acknowledgement_type`, `passive_checks_enabled`, `active_checks_enabled`, `event_handler_enabled`, `flap_detection_enabled`, `is_flapping`, `latency`, `execution_time`, `scheduled_downtime_depth`, `process_performance_data`, `obsess_over_host`, `normal_check_interval`, `retry_check_interval`, `check_timeperiod`, `node_name`, `last_time_up`, `last_time_down`, `last_time_unreachable`, `current_notification_number`, `percent_state_change`, `event_handler`, `check_command`) VALUES
('hplj2605dn', 1538938785, '(Host check timed out after 30.00 seconds)', '', '', 1, 10, 10, 1538933008, 1538938929, 1, 1538769805, 1538769805, 1, 1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.164, 30.001, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538757644, 1538933038, 0, 0, 0, NULL, 'check-host-alive'),
('linksys-srw224p', 1538938785, '(Host check timed out after 30.00 seconds)', '', '', 1, 10, 10, 1538933014, 1538938952, 1, 1538769757, 1538769757, 1, 1, 1538932108, 1538933908, 1, 0, 0, 1, 1, 1, 1, 0, 0.773, 30.001, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538757461, 1538933044, 0, 17, 0, NULL, 'check-host-alive'),
('localhost', 1538938785, 'PING OK - Packet loss = 0%, RTA = 0.14 ms', '', 'rta=0.144000ms;3000.000000;5000.000000;0.000000 pl=0%;80;100;0', 0, 1, 10, 1538933030, 1538938858, 1, 1538926971, 1538926971, 0, 1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.173, 4.043, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538933034, 1538926971, 0, 0, 0, NULL, 'check-host-alive'),
('winserver', 1538938785, 'PING OK - Packet loss = 0%, RTA = 0.33 ms', '', 'rta=0.330000ms;3000.000000;5000.000000;0.000000 pl=0%;80;100;0', 0, 1, 10, 1538933099, 1538938842, 1, 1538758001, 1538758001, 0, 1, 1538758001, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.157, 4.036, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538933103, 1538758001, 0, 0, 0, NULL, 'check-host-alive');

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_host_acknowledgements`
--

CREATE TABLE `statusengine_host_acknowledgements` (
  `hostname` varchar(255) DEFAULT NULL,
  `state` tinyint(2) UNSIGNED DEFAULT '0',
  `author_name` varchar(255) DEFAULT NULL,
  `comment_data` varchar(1024) DEFAULT NULL,
  `entry_time` bigint(13) NOT NULL,
  `acknowledgement_type` tinyint(2) UNSIGNED DEFAULT '0',
  `is_sticky` tinyint(1) UNSIGNED DEFAULT '0',
  `persistent_comment` tinyint(1) UNSIGNED DEFAULT '0',
  `notify_contacts` tinyint(1) UNSIGNED DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_host_downtimehistory`
--

CREATE TABLE `statusengine_host_downtimehistory` (
  `hostname` varchar(255) NOT NULL,
  `entry_time` bigint(13) NOT NULL,
  `author_name` varchar(255) DEFAULT NULL,
  `comment_data` varchar(1024) DEFAULT NULL,
  `internal_downtime_id` int(11) UNSIGNED NOT NULL,
  `triggered_by_id` int(11) UNSIGNED DEFAULT NULL,
  `is_fixed` tinyint(1) UNSIGNED DEFAULT '0',
  `duration` int(11) UNSIGNED DEFAULT NULL,
  `scheduled_start_time` bigint(13) NOT NULL,
  `scheduled_end_time` bigint(13) NOT NULL,
  `was_started` tinyint(1) UNSIGNED DEFAULT '0',
  `actual_start_time` bigint(13) NOT NULL,
  `actual_end_time` bigint(13) NOT NULL,
  `was_cancelled` tinyint(1) UNSIGNED DEFAULT '0',
  `node_name` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_host_notifications`
--

CREATE TABLE `statusengine_host_notifications` (
  `hostname` varchar(255) DEFAULT NULL,
  `contact_name` varchar(1024) DEFAULT NULL,
  `command_name` varchar(1024) DEFAULT NULL,
  `command_args` varchar(1024) DEFAULT NULL,
  `state` tinyint(2) UNSIGNED DEFAULT '0',
  `start_time` bigint(13) NOT NULL,
  `end_time` bigint(13) NOT NULL,
  `reason_type` tinyint(3) UNSIGNED DEFAULT '0',
  `output` varchar(1024) DEFAULT NULL,
  `ack_author` varchar(255) DEFAULT NULL,
  `ack_data` varchar(1024) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_host_scheduleddowntimes`
--

CREATE TABLE `statusengine_host_scheduleddowntimes` (
  `hostname` varchar(255) NOT NULL,
  `entry_time` bigint(13) NOT NULL,
  `author_name` varchar(255) DEFAULT NULL,
  `comment_data` varchar(1024) DEFAULT NULL,
  `internal_downtime_id` int(11) UNSIGNED NOT NULL,
  `triggered_by_id` int(11) UNSIGNED DEFAULT NULL,
  `is_fixed` tinyint(1) UNSIGNED DEFAULT '0',
  `duration` int(11) UNSIGNED DEFAULT NULL,
  `scheduled_start_time` bigint(13) NOT NULL,
  `scheduled_end_time` bigint(13) NOT NULL,
  `was_started` tinyint(1) UNSIGNED DEFAULT '0',
  `actual_start_time` bigint(13) NOT NULL,
  `node_name` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_host_statehistory`
--

CREATE TABLE `statusengine_host_statehistory` (
  `hostname` varchar(255) DEFAULT NULL,
  `state_time` bigint(13) NOT NULL,
  `state_change` tinyint(1) DEFAULT '0',
  `state` tinyint(2) UNSIGNED DEFAULT '0',
  `is_hardstate` tinyint(1) UNSIGNED DEFAULT '0',
  `current_check_attempt` tinyint(3) UNSIGNED DEFAULT '0',
  `max_check_attempts` tinyint(3) UNSIGNED DEFAULT '0',
  `last_state` tinyint(2) UNSIGNED DEFAULT '0',
  `last_hard_state` tinyint(2) UNSIGNED DEFAULT '0',
  `output` varchar(1024) DEFAULT NULL,
  `long_output` varchar(8192) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_logentries`
--

CREATE TABLE `statusengine_logentries` (
  `entry_time` bigint(13) NOT NULL,
  `logentry_type` int(11) DEFAULT '0',
  `logentry_data` varchar(255) DEFAULT NULL,
  `node_name` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Daten für Tabelle `statusengine_logentries`
--

INSERT INTO `statusengine_logentries` (`entry_time`, `logentry_type`, `logentry_data`, `node_name`) VALUES
(1538938770, 262144, 'statusengine: Flush 1 records as bunch to queue statusngin_servicestatus', 'WSL'),
(1538938771, 262144, 'statusengine: Flush 11 records as bunch to queue statusngin_servicestatus', 'WSL'),
(1538938778, 262144, 'statusengine: successfull reconnected to gearmand 127.0.0.1:4730 (23 lost jobs)', 'WSL'),
(1538938778, 262144, 'statusengine: Flush 11 records as bunch to queue statusngin_hoststatus', 'WSL'),
(1538938778, 262144, 'statusengine: Flush 11 records as bunch to queue statusngin_servicestatus', 'WSL'),
(1538938778, 262144, 'statusengine: Flush 11 records as bunch to queue statusngin_servicestatus', 'WSL'),
(1538938778, 262144, 'statusengine: Flush 11 records as bunch to queue statusngin_servicestatus', 'WSL'),
(1538938778, 262144, 'statusengine: Flush 11 records as bunch to queue statusngin_servicestatus', 'WSL'),
(1538938778, 262144, 'Successfully launched command file worker with pid 90', 'WSL'),
(1538938784, 64, 'Caught \'Interrupt\', shutting down...', 'WSL'),
(1538938784, 262144, 'statusengine: Flush 7 records as bunch to queue statusngin_servicestatus', 'WSL'),
(1538938785, 262144, 'statusengine: Flush 1 records as bunch to queue statusngin_objects', 'WSL'),
(1538938785, 262144, 'statusengine: Flush 11 records as bunch to queue statusngin_servicestatus', 'WSL'),
(1538938785, 262144, 'statusengine: Flush 11 records as bunch to queue statusngin_hoststatus', 'WSL'),
(1538938785, 262144, 'statusengine: Flush 11 records as bunch to queue statusngin_servicestatus', 'WSL'),
(1538938785, 262144, 'statusengine: Flush 11 records as bunch to queue statusngin_servicestatus', 'WSL'),
(1538938785, 262144, 'statusengine: Flush 11 records as bunch to queue statusngin_servicestatus', 'WSL'),
(1538938785, 262144, 'statusengine: Flush 11 records as bunch to queue statusngin_servicestatus', 'WSL'),
(1538938785, 262144, 'Successfully launched command file worker with pid 120', 'WSL'),
(1538938792, 64, 'Caught \'Interrupt\', shutting down...', 'WSL'),
(1538938792, 262144, 'statusengine: Flush 8 records as bunch to queue statusngin_servicestatus', 'WSL');

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_nodes`
--

CREATE TABLE `statusengine_nodes` (
  `node_name` varchar(255) NOT NULL,
  `node_version` varchar(255) DEFAULT NULL,
  `node_start_time` bigint(13) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Daten für Tabelle `statusengine_nodes`
--

INSERT INTO `statusengine_nodes` (`node_name`, `node_version`, `node_start_time`) VALUES
('WSL', '3.2.0', 1538938876);

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_perfdata`
--

CREATE TABLE `statusengine_perfdata` (
  `hostname` varchar(255) DEFAULT NULL,
  `service_description` varchar(255) DEFAULT NULL,
  `label` varchar(255) DEFAULT NULL,
  `timestamp` bigint(20) NOT NULL,
  `timestamp_unix` bigint(13) NOT NULL,
  `value` double DEFAULT NULL,
  `unit` varchar(10) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_servicechecks`
--

CREATE TABLE `statusengine_servicechecks` (
  `hostname` varchar(255) DEFAULT NULL,
  `service_description` varchar(255) DEFAULT NULL,
  `state` tinyint(2) UNSIGNED DEFAULT '0',
  `is_hardstate` tinyint(1) UNSIGNED DEFAULT '0',
  `start_time` bigint(13) NOT NULL,
  `end_time` bigint(13) NOT NULL,
  `output` varchar(1024) DEFAULT NULL,
  `timeout` tinyint(3) UNSIGNED DEFAULT '0',
  `early_timeout` tinyint(1) UNSIGNED DEFAULT '0',
  `latency` float DEFAULT '0',
  `execution_time` float DEFAULT '0',
  `perfdata` varchar(1024) DEFAULT NULL,
  `command` varchar(1024) DEFAULT NULL,
  `current_check_attempt` tinyint(3) UNSIGNED DEFAULT '0',
  `max_check_attempts` tinyint(3) UNSIGNED DEFAULT '0',
  `long_output` varchar(8192) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_servicestatus`
--

CREATE TABLE `statusengine_servicestatus` (
  `hostname` varchar(255) NOT NULL,
  `service_description` varchar(255) NOT NULL,
  `status_update_time` bigint(13) NOT NULL,
  `output` varchar(1024) DEFAULT NULL,
  `long_output` varchar(1024) DEFAULT NULL,
  `perfdata` varchar(1024) DEFAULT NULL,
  `current_state` tinyint(2) UNSIGNED DEFAULT '0',
  `current_check_attempt` tinyint(3) UNSIGNED DEFAULT '0',
  `max_check_attempts` tinyint(3) UNSIGNED DEFAULT '0',
  `last_check` bigint(13) NOT NULL,
  `next_check` bigint(13) NOT NULL,
  `is_passive_check` tinyint(1) UNSIGNED DEFAULT '0',
  `last_state_change` bigint(13) NOT NULL,
  `last_hard_state_change` bigint(13) NOT NULL,
  `last_hard_state` tinyint(2) UNSIGNED DEFAULT '0',
  `is_hardstate` tinyint(1) UNSIGNED DEFAULT '0',
  `last_notification` bigint(13) NOT NULL,
  `next_notification` bigint(13) NOT NULL,
  `notifications_enabled` tinyint(1) UNSIGNED DEFAULT '0',
  `problem_has_been_acknowledged` tinyint(1) UNSIGNED DEFAULT '0',
  `acknowledgement_type` tinyint(1) UNSIGNED DEFAULT '0',
  `passive_checks_enabled` tinyint(1) UNSIGNED DEFAULT '0',
  `active_checks_enabled` tinyint(1) UNSIGNED DEFAULT '0',
  `event_handler_enabled` tinyint(1) UNSIGNED DEFAULT '0',
  `flap_detection_enabled` tinyint(1) UNSIGNED DEFAULT '0',
  `is_flapping` tinyint(1) UNSIGNED DEFAULT '0',
  `latency` float DEFAULT '0',
  `execution_time` float DEFAULT '0',
  `scheduled_downtime_depth` tinyint(2) UNSIGNED DEFAULT '0',
  `process_performance_data` tinyint(1) UNSIGNED DEFAULT '0',
  `obsess_over_service` tinyint(1) UNSIGNED DEFAULT '0',
  `normal_check_interval` int(11) UNSIGNED DEFAULT '0',
  `retry_check_interval` int(11) UNSIGNED DEFAULT '0',
  `check_timeperiod` varchar(255) DEFAULT NULL,
  `node_name` varchar(255) DEFAULT NULL,
  `last_time_ok` bigint(13) NOT NULL,
  `last_time_warning` bigint(13) NOT NULL,
  `last_time_critical` bigint(13) NOT NULL,
  `last_time_unknown` bigint(13) NOT NULL,
  `current_notification_number` int(11) UNSIGNED DEFAULT '0',
  `percent_state_change` double DEFAULT '0',
  `event_handler` varchar(255) DEFAULT NULL,
  `check_command` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Daten für Tabelle `statusengine_servicestatus`
--

INSERT INTO `statusengine_servicestatus` (`hostname`, `service_description`, `status_update_time`, `output`, `long_output`, `perfdata`, `current_state`, `current_check_attempt`, `max_check_attempts`, `last_check`, `next_check`, `is_passive_check`, `last_state_change`, `last_hard_state_change`, `last_hard_state`, `is_hardstate`, `last_notification`, `next_notification`, `notifications_enabled`, `problem_has_been_acknowledged`, `acknowledgement_type`, `passive_checks_enabled`, `active_checks_enabled`, `event_handler_enabled`, `flap_detection_enabled`, `is_flapping`, `latency`, `execution_time`, `scheduled_downtime_depth`, `process_performance_data`, `obsess_over_service`, `normal_check_interval`, `retry_check_interval`, `check_timeperiod`, `node_name`, `last_time_ok`, `last_time_warning`, `last_time_critical`, `last_time_unknown`, `current_notification_number`, `percent_state_change`, `event_handler`, `check_command`) VALUES
('hplj2605dn', 'PING', 1538938785, 'CRITICAL - Plugin timed out after 30 seconds', '', '', 2, 3, 3, 1538932977, 1538939137, 1, 1538757812, 1538757812, 2, 1, 0, 0, 1, 1, 2, 1, 1, 1, 1, 0, 0.138, 31.026, 0, 1, 1, 10, 1, '24x7', 'WSL', 1538757812, 0, 1538932977, 0, 0, 0, NULL, 'check_ping!3000.0,80%!5000.0,100%'),
('hplj2605dn', 'Printer Status', 1538938785, 'Timeout: No Response from 192.168.1.30:161. : Timeout from host 192.168.1.30', '\\n', '', 2, 3, 3, 1538932761, 1538939093, 1, 1538930842, 1538930842, 2, 1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.136, 23.629, 0, 1, 1, 10, 1, '24x7', 'WSL', 1538930842, 0, 1538932761, 0, 0, 12.11, NULL, 'check_hpjd!-C public'),
('linksys-srw224p', 'PING', 1538938785, 'PING CRITICAL - Packet loss = 100%', '', 'rta=600.000000ms;200.000000;600.000000;0.000000 pl=100%;20;60;0', 2, 3, 3, 1538933004, 1538938880, 1, 1538757573, 1538757573, 2, 1, 0, 0, 1, 1, 2, 1, 1, 1, 1, 0, 0.139, 10.628, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538757573, 0, 1538933004, 0, 0, 0, NULL, 'check_ping!200.0,20%!600.0,60%'),
('linksys-srw224p', 'Port 1 Bandwidth Usage', 1538938785, 'check_mrtgtraf: Unable to open MRTG log file', 'Usage check_mrtgtraf -F <log_file> -a <AVG', 'MAX> -w <warning_pair> -c <critical_pair> [-e expire_minutes]', 3, 3, 3, 1538932913, 1538939009, 1, 1538926998, 1538926998, 3, 1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.137, 0.037, 0, 1, 1, 10, 2, '24x7', 'WSL', 1538926998, 0, 1538758063, 1538932913, 0, 9.47, NULL, 'check_local_mrtgtraf!/var/lib/mrtg/192.168.1.253_1.log!AVG!1000000,1000000!5000000,5000000!10'),
('linksys-srw224p', 'Port 1 Link Status', 1538938785, 'External command error: MIB search path: /home/nook/.snmp/mibs:/usr/share/snmp/mibs:/usr/share/snmp/mibs/iana:/usr/share/snmp/mibs/ietf:/usr/share/mibs/site:/usr/share/snmp/mibs:/usr/share/mibs/iana:/usr/share/mibs/ietf:/usr/share/mibs/netsnmp', 'Cannot find module (SNMPv2-SMI): At line 1 in (none)\\nifOperStatus.1: Unknown Object Identifier (Sub-id not found: (top) -> ifOperStatus)\\n', '', 3, 3, 3, 1538932912, 1538939188, 1, 1538768868, 1538768868, 3, 1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.137, 0.034, 0, 1, 1, 10, 2, '24x7', 'WSL', 1538768868, 0, 0, 1538932912, 0, 0, NULL, 'check_snmp!-C public -o ifOperStatus.1 -r 1 -m RFC1213-MIB'),
('linksys-srw224p', 'Uptime', 1538938785, 'CRITICAL - Plugin timed out while executing system call', '', '', 2, 3, 3, 1538932543, 1538939179, 1, 1538758036, 1538758036, 2, 1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.136, 55.029, 0, 1, 1, 10, 2, '24x7', 'WSL', 1538758036, 0, 1538932543, 0, 0, 0, NULL, 'check_snmp!-C public -o sysUpTime.0'),
('localhost', 'Current Load', 1538938785, 'OK - load average: 0.52, 0.58, 0.59', '', 'load1=0.520;5.000;10.000;0; load5=0.580;4.000;6.000;0; load15=0.590;3.000;4.000;0;', 0, 1, 4, 1538932999, 1538938948, 1, 1538930780, 1538758167, 0, 1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.139, 0.049, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538932999, 1538925128, 1538930780, 0, 0, 0, NULL, 'check_local_load!5.0,4.0,3.0!10.0,6.0,4.0'),
('localhost', 'Current Users', 1538938785, 'USERS OK - 0 users currently logged in', '', 'users=0;20;50;0', 0, 1, 4, 1538933049, 1538939032, 1, 1538758155, 1538758155, 0, 1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.14, 0.043, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538933049, 0, 1538758155, 0, 0, 0, NULL, 'check_local_users!20!50'),
('localhost', 'HTTP', 1538938785, 'connect to address 127.0.0.1 and port 80: Connection refused', 'HTTP CRITICAL - Unable to open TCP socket\\n', '', 2, 4, 4, 1538932909, 1538939026, 1, 1538771730, 1538771730, 2, 1, 0, 0, 0, 1, 2, 1, 1, 1, 1, 0, 0.139, 1.03, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538757485, 1538771730, 1538932909, 0, 0, 0, NULL, 'check_http!-u /naemon/'),
('localhost', 'PING', 1538938785, 'PING OK - Packet loss = 0%, RTA = 0.21 ms', '', 'rta=0.208000ms;100.000000;500.000000;0.000000 pl=0%;20;60;0', 0, 1, 4, 1538933055, 1538939083, 1, 1538758162, 1538758162, 0, 1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.138, 4.037, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538933055, 0, 1538758162, 0, 0, 0, NULL, 'check_ping!100.0,20%!500.0,60%'),
('localhost', 'Root Partition', 1538938785, 'DISK OK - free space: / 79628 MB (33% inode=-1846520928299154432%):', '', '/=158280MB;190326;214117;0;237908', 0, 1, 4, 1538932877, 1538938924, 1, 1538768182, 1538768182, 0, 1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.138, 0.034, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538932877, 0, 1538768182, 0, 0, 0, NULL, 'check_local_disk!20%!10%!/'),
('localhost', 'SSH', 1538938785, 'connect to address 127.0.0.1 and port 22: Connection refused', '', '', 2, 4, 4, 1538933029, 1538938939, 1, 1538757533, 1538757533, 2, 1, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0.139, 1.03, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538757533, 0, 1538933029, 0, 0, 0, NULL, 'check_ssh'),
('localhost', 'Swap Usage', 1538938785, 'SWAP OK - 100% free (29709 MB out of 29738 MB)', '', 'swap=29709MB;0;0;0;29738', 0, 1, 4, 1538932858, 1538939030, 1, 1538768022, 1538768022, 0, 1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.139, 0.039, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538932858, 0, 1538768022, 0, 0, 0, NULL, 'check_local_swap!20!10'),
('localhost', 'Total Processes', 1538938785, 'PROCS OK: 17 processes with STATE = RSZDT', '', 'procs=17;250;400;0;', 0, 1, 4, 1538932948, 1538939016, 1, 1538768841, 1538768841, 0, 1, 0, 0, 1, 0, 0, 1, 1, 1, 1, 0, 0.14, 0.025, 0, 1, 1, 5, 1, '24x7', 'WSL', 1538932948, 0, 1538768841, 0, 0, 0, NULL, 'check_local_procs!250!400!RSZDT'),
('winserver', 'C:\\ Drive Space', 1538938785, 'connect to address 192.168.1.2 and port 12489: Connection refused', 'could not fetch information from server\\n', '', 2, 3, 3, 1538933098, 1538938973, 1, 1538757566, 1538757566, 2, 1, 1538930629, 1538934229, 1, 0, 0, 1, 1, 1, 1, 0, 0.138, 1.017, 0, 1, 1, 10, 2, '24x7', 'WSL', 1538757446, 0, 1538933098, 0, 10, 0, NULL, 'check_nt!USEDDISKSPACE!-l c -w 80 -c 90'),
('winserver', 'CPU Load', 1538938785, 'connect to address 192.168.1.2 and port 12489: Connection refused', 'could not fetch information from server\\n', '', 2, 3, 3, 1538932792, 1538939380, 1, 1538757817, 1538757817, 2, 1, 1538930130, 1538933730, 1, 0, 0, 1, 1, 1, 1, 0, 0.138, 1.03, 0, 1, 1, 10, 2, '24x7', 'WSL', 1538757817, 0, 1538932792, 0, 9, 0, NULL, 'check_nt!CPULOAD!-l 5,80,90'),
('winserver', 'Explorer', 1538938785, 'connect to address 192.168.1.2 and port 12489: Connection refused', 'could not fetch information from server\\n', '', 2, 3, 3, 1538932769, 1538939000, 1, 1538757648, 1538757648, 2, 1, 1538930900, 1538934500, 1, 0, 0, 1, 1, 1, 1, 0, 0.137, 1.019, 0, 1, 1, 10, 2, '24x7', 'WSL', 1538757648, 0, 1538932769, 0, 9, 0, NULL, 'check_nt!PROCSTATE!-d SHOWALL -l Explorer.exe'),
('winserver', 'Memory Usage', 1538938785, 'connect to address 192.168.1.2 and port 12489: Connection refused', 'could not fetch information from server\\n', '', 2, 3, 3, 1538932974, 1538939158, 1, 1538757877, 1538757877, 2, 1, 1538930041, 1538933641, 1, 0, 0, 1, 1, 1, 1, 0, 0.139, 1.034, 0, 1, 1, 10, 2, '24x7', 'WSL', 1538757877, 0, 1538932974, 0, 9, 0, NULL, 'check_nt!MEMUSE!-w 80 -c 90'),
('winserver', 'NSClient++ Version', 1538938785, 'connect to address 192.168.1.2 and port 12489: Connection refused', 'could not fetch information from server\\n', '', 2, 3, 3, 1538932718, 1538939057, 1, 1538757767, 1538757767, 2, 1, 1538930191, 1538933791, 1, 0, 0, 1, 1, 1, 1, 0, 0.137, 1.012, 0, 1, 1, 10, 2, '24x7', 'WSL', 1538757767, 0, 1538932718, 0, 9, 0, NULL, 'check_nt!CLIENTVERSION'),
('winserver', 'Uptime', 1538938785, 'connect to address 192.168.1.2 and port 12489: Connection refused', 'could not fetch information from server\\n', '', 2, 3, 3, 1538932752, 1538939164, 1, 1538757522, 1538757522, 2, 1, 1538930361, 1538933961, 1, 0, 0, 1, 1, 1, 1, 0, 0.137, 1.018, 0, 1, 1, 10, 2, '24x7', 'WSL', 1538757522, 0, 1538932752, 0, 9, 0, NULL, 'check_nt!UPTIME'),
('winserver', 'W3SVC', 1538938785, 'connect to address 192.168.1.2 and port 12489: Connection refused', 'could not fetch information from server\\n', '', 2, 3, 3, 1538932577, 1538939317, 1, 1538757587, 1538757587, 2, 1, 1538930473, 1538934073, 1, 0, 0, 1, 1, 1, 1, 0, 0.138, 1.022, 0, 1, 1, 10, 2, '24x7', 'WSL', 1538757587, 0, 1538932577, 0, 10, 0, NULL, 'check_nt!SERVICESTATE!-d SHOWALL -l W3SVC');

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_service_acknowledgements`
--

CREATE TABLE `statusengine_service_acknowledgements` (
  `hostname` varchar(255) DEFAULT NULL,
  `service_description` varchar(255) DEFAULT NULL,
  `state` tinyint(2) UNSIGNED DEFAULT '0',
  `author_name` varchar(255) DEFAULT NULL,
  `comment_data` varchar(1024) DEFAULT NULL,
  `entry_time` bigint(13) NOT NULL,
  `acknowledgement_type` tinyint(2) UNSIGNED DEFAULT '0',
  `is_sticky` tinyint(1) UNSIGNED DEFAULT '0',
  `persistent_comment` tinyint(1) UNSIGNED DEFAULT '0',
  `notify_contacts` tinyint(1) UNSIGNED DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_service_downtimehistory`
--

CREATE TABLE `statusengine_service_downtimehistory` (
  `hostname` varchar(255) NOT NULL,
  `service_description` varchar(255) NOT NULL,
  `entry_time` bigint(13) NOT NULL,
  `author_name` varchar(255) DEFAULT NULL,
  `comment_data` varchar(1024) DEFAULT NULL,
  `internal_downtime_id` int(11) UNSIGNED NOT NULL,
  `triggered_by_id` int(11) UNSIGNED DEFAULT NULL,
  `is_fixed` tinyint(1) UNSIGNED DEFAULT '0',
  `duration` int(11) UNSIGNED DEFAULT NULL,
  `scheduled_start_time` bigint(13) NOT NULL,
  `scheduled_end_time` bigint(13) NOT NULL,
  `was_started` tinyint(1) UNSIGNED DEFAULT '0',
  `actual_start_time` bigint(13) NOT NULL,
  `actual_end_time` bigint(13) NOT NULL,
  `was_cancelled` tinyint(1) UNSIGNED DEFAULT '0',
  `node_name` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_service_notifications`
--

CREATE TABLE `statusengine_service_notifications` (
  `hostname` varchar(255) DEFAULT NULL,
  `service_description` varchar(255) DEFAULT NULL,
  `contact_name` varchar(1024) DEFAULT NULL,
  `command_name` varchar(1024) DEFAULT NULL,
  `command_args` varchar(1024) DEFAULT NULL,
  `state` tinyint(2) UNSIGNED DEFAULT '0',
  `start_time` bigint(13) NOT NULL,
  `end_time` bigint(13) NOT NULL,
  `reason_type` tinyint(3) UNSIGNED DEFAULT '0',
  `output` varchar(1024) DEFAULT NULL,
  `ack_author` varchar(255) DEFAULT NULL,
  `ack_data` varchar(1024) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_service_scheduleddowntimes`
--

CREATE TABLE `statusengine_service_scheduleddowntimes` (
  `hostname` varchar(255) NOT NULL,
  `service_description` varchar(255) NOT NULL,
  `entry_time` bigint(13) NOT NULL,
  `author_name` varchar(255) DEFAULT NULL,
  `comment_data` varchar(1024) DEFAULT NULL,
  `internal_downtime_id` int(11) UNSIGNED NOT NULL,
  `triggered_by_id` int(11) UNSIGNED DEFAULT NULL,
  `is_fixed` tinyint(1) UNSIGNED DEFAULT '0',
  `duration` int(11) UNSIGNED DEFAULT NULL,
  `scheduled_start_time` bigint(13) NOT NULL,
  `scheduled_end_time` bigint(13) NOT NULL,
  `was_started` tinyint(1) UNSIGNED DEFAULT '0',
  `actual_start_time` bigint(13) NOT NULL,
  `node_name` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_service_statehistory`
--

CREATE TABLE `statusengine_service_statehistory` (
  `hostname` varchar(255) DEFAULT NULL,
  `service_description` varchar(255) DEFAULT NULL,
  `state_time` bigint(13) NOT NULL,
  `state_change` tinyint(1) DEFAULT '0',
  `state` tinyint(2) UNSIGNED DEFAULT '0',
  `is_hardstate` tinyint(1) UNSIGNED DEFAULT '0',
  `current_check_attempt` tinyint(3) UNSIGNED DEFAULT '0',
  `max_check_attempts` tinyint(3) UNSIGNED DEFAULT '0',
  `last_state` tinyint(2) UNSIGNED DEFAULT '0',
  `last_hard_state` tinyint(2) UNSIGNED DEFAULT '0',
  `output` varchar(1024) DEFAULT NULL,
  `long_output` varchar(8192) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_tasks`
--

CREATE TABLE `statusengine_tasks` (
  `uuid` varchar(255) DEFAULT NULL,
  `node_name` varchar(255) DEFAULT NULL,
  `entry_time` bigint(13) NOT NULL,
  `type` varchar(255) DEFAULT NULL,
  `payload` varchar(8192) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- Tabellenstruktur für Tabelle `statusengine_users`
--

CREATE TABLE `statusengine_users` (
  `username` varchar(255) DEFAULT NULL,
  `password` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- Indizes der exportierten Tabellen
--

--
-- Indizes für die Tabelle `statusengine_dbversion`
--
ALTER TABLE `statusengine_dbversion`
  ADD PRIMARY KEY (`id`);

--
-- Indizes für die Tabelle `statusengine_hostchecks`
--
ALTER TABLE `statusengine_hostchecks`
  ADD KEY `times` (`start_time`,`end_time`),
  ADD KEY `hostname` (`hostname`,`start_time`);

--
-- Indizes für die Tabelle `statusengine_hoststatus`
--
ALTER TABLE `statusengine_hoststatus`
  ADD PRIMARY KEY (`hostname`);

--
-- Indizes für die Tabelle `statusengine_host_acknowledgements`
--
ALTER TABLE `statusengine_host_acknowledgements`
  ADD KEY `hostname` (`hostname`),
  ADD KEY `entry_time` (`entry_time`);

--
-- Indizes für die Tabelle `statusengine_host_downtimehistory`
--
ALTER TABLE `statusengine_host_downtimehistory`
  ADD PRIMARY KEY (`hostname`,`node_name`,`scheduled_start_time`,`internal_downtime_id`),
  ADD KEY `list` (`hostname`,`scheduled_start_time`,`scheduled_end_time`,`was_cancelled`);

--
-- Indizes für die Tabelle `statusengine_host_notifications`
--
ALTER TABLE `statusengine_host_notifications`
  ADD KEY `hostname` (`hostname`),
  ADD KEY `start_time` (`start_time`);

--
-- Indizes für die Tabelle `statusengine_host_scheduleddowntimes`
--
ALTER TABLE `statusengine_host_scheduleddowntimes`
  ADD PRIMARY KEY (`hostname`,`node_name`,`scheduled_start_time`,`internal_downtime_id`);

--
-- Indizes für die Tabelle `statusengine_host_statehistory`
--
ALTER TABLE `statusengine_host_statehistory`
  ADD KEY `hostname_time` (`hostname`,`state_time`);

--
-- Indizes für die Tabelle `statusengine_logentries`
--
ALTER TABLE `statusengine_logentries`
  ADD KEY `logentries` (`entry_time`,`logentry_data`,`node_name`),
  ADD KEY `logentry_data_time` (`logentry_data`,`entry_time`);

--
-- Indizes für die Tabelle `statusengine_nodes`
--
ALTER TABLE `statusengine_nodes`
  ADD PRIMARY KEY (`node_name`);

--
-- Indizes für die Tabelle `statusengine_perfdata`
--
ALTER TABLE `statusengine_perfdata`
  ADD KEY `metric` (`hostname`,`service_description`,`label`,`timestamp_unix`),
  ADD KEY `timestamp_unix` (`timestamp_unix`);

--
-- Indizes für die Tabelle `statusengine_servicechecks`
--
ALTER TABLE `statusengine_servicechecks`
  ADD KEY `servicename` (`hostname`,`service_description`,`start_time`);

--
-- Indizes für die Tabelle `statusengine_servicestatus`
--
ALTER TABLE `statusengine_servicestatus`
  ADD PRIMARY KEY (`hostname`,`service_description`),
  ADD KEY `current_state_node` (`current_state`,`node_name`),
  ADD KEY `issues` (`problem_has_been_acknowledged`,`scheduled_downtime_depth`,`current_state`),
  ADD KEY `service_description` (`service_description`);

--
-- Indizes für die Tabelle `statusengine_service_acknowledgements`
--
ALTER TABLE `statusengine_service_acknowledgements`
  ADD KEY `servicename` (`hostname`,`service_description`),
  ADD KEY `entry_time` (`entry_time`),
  ADD KEY `servicedesc_time` (`service_description`,`entry_time`);

--
-- Indizes für die Tabelle `statusengine_service_downtimehistory`
--
ALTER TABLE `statusengine_service_downtimehistory`
  ADD PRIMARY KEY (`hostname`,`service_description`,`node_name`,`scheduled_start_time`,`internal_downtime_id`),
  ADD KEY `report` (`service_description`,`scheduled_start_time`,`scheduled_end_time`,`was_cancelled`);

--
-- Indizes für die Tabelle `statusengine_service_notifications`
--
ALTER TABLE `statusengine_service_notifications`
  ADD KEY `servicename` (`hostname`,`service_description`),
  ADD KEY `start_time` (`start_time`);

--
-- Indizes für die Tabelle `statusengine_service_scheduleddowntimes`
--
ALTER TABLE `statusengine_service_scheduleddowntimes`
  ADD PRIMARY KEY (`hostname`,`service_description`,`node_name`,`scheduled_start_time`,`internal_downtime_id`);

--
-- Indizes für die Tabelle `statusengine_service_statehistory`
--
ALTER TABLE `statusengine_service_statehistory`
  ADD KEY `host_servicename_time` (`hostname`,`service_description`,`state_time`),
  ADD KEY `servicename_time` (`service_description`,`state_time`);

--
-- Indizes für die Tabelle `statusengine_tasks`
--
ALTER TABLE `statusengine_tasks`
  ADD KEY `node_name` (`node_name`),
  ADD KEY `uuid` (`uuid`);

--
-- Indizes für die Tabelle `statusengine_users`
--
ALTER TABLE `statusengine_users`
  ADD KEY `username` (`username`,`password`);

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
