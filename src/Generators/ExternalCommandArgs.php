<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2018  Daniel Ziegler
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\Generators;


use Statusengine\Exceptions\ExternalCommandNotSupportedException;
use Statusengine\ValueObjects\ExternalCommandArgsQueryOptions;

class ExternalCommandArgs {

    private $supportedCommands = [
        'PROCESS_HOST_CHECK_RESULT',
        'PROCESS_SERVICE_CHECK_RESULT',
        'SEND_CUSTOM_HOST_NOTIFICATION',
        'SEND_CUSTOM_SVC_NOTIFICATION',
        'ACKNOWLEDGE_HOST_PROBLEM',
        'ACKNOWLEDGE_SVC_PROBLEM',
        'SCHEDULE_HOST_DOWNTIME',
        'SCHEDULE_HOST_SVC_DOWNTIME',
        'SCHEDULE_AND_PROPAGATE_TRIGGERED_HOST_DOWNTIME',
        'SCHEDULE_AND_PROPAGATE_HOST_DOWNTIME',
        'SCHEDULE_SVC_DOWNTIME',
        'DEL_HOST_DOWNTIME',
        'DEL_SVC_DOWNTIME'
    ];

    /**
     * @var ExternalCommandArgsQueryOptions
     */
    private $ExternalCommandArgsQueryOptions;

    /**
     * ExternalCommandArgs constructor.
     * @param ExternalCommandArgsQueryOptions $ExternalCommandArgsQueryOptions
     */
    public function __construct(ExternalCommandArgsQueryOptions $ExternalCommandArgsQueryOptions) {
        $this->ExternalCommandArgsQueryOptions = $ExternalCommandArgsQueryOptions;
        $this->isCommandSupported();
    }

    /**
     * @return bool
     * @throws ExternalCommandNotSupportedException
     */
    public function isCommandSupported() {
        if (!in_array($this->ExternalCommandArgsQueryOptions->getCommandName(), $this->supportedCommands)) {
            throw new ExternalCommandNotSupportedException(sprintf(
                'External command %s is not supported',
                $this->ExternalCommandArgsQueryOptions->getCommandName()
            ));
        }
        return true;
    }

    /**
     * @return string|array
     */
    public function getCommand() {
        switch ($this->ExternalCommandArgsQueryOptions->getCommandName()) {
            case 'PROCESS_HOST_CHECK_RESULT':
                return sprintf(
                    '[%s] PROCESS_HOST_CHECK_RESULT;%s;%s;%s',
                    time(),
                    $this->ExternalCommandArgsQueryOptions->getHostname(),
                    $this->ExternalCommandArgsQueryOptions->getState(),
                    $this->ExternalCommandArgsQueryOptions->getOutput()
                );
                break;

            case 'PROCESS_SERVICE_CHECK_RESULT':
                return sprintf(
                    '[%s] PROCESS_SERVICE_CHECK_RESULT;%s;%s;%s;%s',
                    time(),
                    $this->ExternalCommandArgsQueryOptions->getHostname(),
                    $this->ExternalCommandArgsQueryOptions->getServicedescription(),
                    $this->ExternalCommandArgsQueryOptions->getState(),
                    $this->ExternalCommandArgsQueryOptions->getOutput()
                );
                break;

            case 'SEND_CUSTOM_HOST_NOTIFICATION':
                $options = 0;
                if ($this->ExternalCommandArgsQueryOptions->isBroadcast()) {
                    $options = 1;
                }
                if ($this->ExternalCommandArgsQueryOptions->isForce()) {
                    $options += 2;
                }

                return sprintf(
                    '[%s] SEND_CUSTOM_HOST_NOTIFICATION;%s;%s;%s;%s',
                    time(),
                    $this->ExternalCommandArgsQueryOptions->getHostname(),
                    $options,
                    $this->ExternalCommandArgsQueryOptions->getAuthorName(),
                    $this->ExternalCommandArgsQueryOptions->getComment()
                );
                break;

            case 'SEND_CUSTOM_SVC_NOTIFICATION':
                $options = 0;
                if ($this->ExternalCommandArgsQueryOptions->isBroadcast()) {
                    $options = 1;
                }
                if ($this->ExternalCommandArgsQueryOptions->isForce()) {
                    $options += 2;
                }

                return sprintf(
                    '[%s] SEND_CUSTOM_SVC_NOTIFICATION;%s;%s;%s;%s;%s',
                    time(),
                    $this->ExternalCommandArgsQueryOptions->getHostname(),
                    $this->ExternalCommandArgsQueryOptions->getServicedescription(),
                    $options,
                    $this->ExternalCommandArgsQueryOptions->getAuthorName(),
                    $this->ExternalCommandArgsQueryOptions->getComment()
                );
                break;

            case 'ACKNOWLEDGE_HOST_PROBLEM':
                $sticky = 0;
                if ($this->ExternalCommandArgsQueryOptions->isSticky()) {
                    $sticky = 2;
                }

                return sprintf(
                    '[%s] ACKNOWLEDGE_HOST_PROBLEM;%s;%s;%s;%s;%s;%s',
                    time(),
                    $this->ExternalCommandArgsQueryOptions->getHostname(),
                    $sticky,
                    1, // notify
                    1, // persistent,
                    $this->ExternalCommandArgsQueryOptions->getAuthorName(),
                    $this->ExternalCommandArgsQueryOptions->getComment()

                );
                break;

            case 'ACKNOWLEDGE_SVC_PROBLEM':
                $sticky = 0;
                if ($this->ExternalCommandArgsQueryOptions->isSticky()) {
                    $sticky = 2;
                }

                return sprintf(
                    '[%s] ACKNOWLEDGE_SVC_PROBLEM;%s;%s;%s;%s;%s;%s;%s',
                    time(),
                    $this->ExternalCommandArgsQueryOptions->getHostname(),
                    $this->ExternalCommandArgsQueryOptions->getServicedescription(),
                    $sticky,
                    1, // notify
                    1, // persistent,
                    $this->ExternalCommandArgsQueryOptions->getAuthorName(),
                    $this->ExternalCommandArgsQueryOptions->getComment()

                );
                break;

            case 'SCHEDULE_HOST_SVC_DOWNTIME':
                $commands = [];

                //<host_name>;<start_time>;<end_time>;<fixed>;<trigger_id>;<duration>;<author>;<comment>
                $commands[] = sprintf(
                    '[%s] SCHEDULE_HOST_DOWNTIME;%s;%s;%s;%s;%s;%s;%s;%s',
                    time(),
                    $this->ExternalCommandArgsQueryOptions->getHostname(),
                    $this->ExternalCommandArgsQueryOptions->getStart(),
                    $this->ExternalCommandArgsQueryOptions->getEnd(),
                    1, //Fixed
                    0, //trigger_id
                    $this->ExternalCommandArgsQueryOptions->getDuration(),
                    $this->ExternalCommandArgsQueryOptions->getAuthorName(),
                    $this->ExternalCommandArgsQueryOptions->getComment()
                );
                $commands[] = sprintf(
                    '[%s] SCHEDULE_HOST_SVC_DOWNTIME;%s;%s;%s;%s;%s;%s;%s;%s',
                    time(),
                    $this->ExternalCommandArgsQueryOptions->getHostname(),
                    $this->ExternalCommandArgsQueryOptions->getStart(),
                    $this->ExternalCommandArgsQueryOptions->getEnd(),
                    1, //Fixed
                    0, //trigger_id
                    $this->ExternalCommandArgsQueryOptions->getDuration(),
                    $this->ExternalCommandArgsQueryOptions->getAuthorName(),
                    $this->ExternalCommandArgsQueryOptions->getComment()
                );
                return $commands;
                break;

            case 'SCHEDULE_HOST_DOWNTIME':
            case 'SCHEDULE_AND_PROPAGATE_TRIGGERED_HOST_DOWNTIME':
            case 'SCHEDULE_AND_PROPAGATE_HOST_DOWNTIME':
                return sprintf(
                    '[%s] %s;%s;%s;%s;%s;%s;%s;%s;%s',
                    time(),
                    $this->ExternalCommandArgsQueryOptions->getCommandName(),
                    $this->ExternalCommandArgsQueryOptions->getHostname(),
                    $this->ExternalCommandArgsQueryOptions->getStart(),
                    $this->ExternalCommandArgsQueryOptions->getEnd(),
                    1, //Fixed
                    0, //trigger_id
                    $this->ExternalCommandArgsQueryOptions->getDuration(),
                    $this->ExternalCommandArgsQueryOptions->getAuthorName(),
                    $this->ExternalCommandArgsQueryOptions->getComment()
                );
                break;

            case 'SCHEDULE_SVC_DOWNTIME':
                return sprintf(
                    '[%s] SCHEDULE_SVC_DOWNTIME;%s;%s;%s;%s;%s;%s;%s;%s;%s',
                    time(),
                    $this->ExternalCommandArgsQueryOptions->getHostname(),
                    $this->ExternalCommandArgsQueryOptions->getServicedescription(),
                    $this->ExternalCommandArgsQueryOptions->getStart(),
                    $this->ExternalCommandArgsQueryOptions->getEnd(),
                    1, //Fixed
                    0, //trigger_id
                    $this->ExternalCommandArgsQueryOptions->getDuration(),
                    $this->ExternalCommandArgsQueryOptions->getAuthorName(),
                    $this->ExternalCommandArgsQueryOptions->getComment()
                );
                break;

            case 'DEL_HOST_DOWNTIME':
            case 'DEL_SVC_DOWNTIME':
            return sprintf(
                '[%s] %s;%s',
                time(),
                $this->ExternalCommandArgsQueryOptions->getCommandName(),
                $this->ExternalCommandArgsQueryOptions->getDowntimeId()
            );


            default:
                return '';
        }
    }

}
