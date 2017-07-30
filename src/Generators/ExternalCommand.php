<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler
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


use Statusengine\ValueObjects\ExternalCommandQueryOptions;

class ExternalCommand {
    const SCHEDULE_FORCED_SVC_CHECK = 1;
    const SCHEDULE_FORCED_HOST_CHECK = 2;
    const SCHEDULE_FORCED_HOST_SVC_CHECKS = 3;
    const DISABLE_SVC_NOTIFICATIONS = 10;
    const ENABLE_SVC_NOTIFICATIONS = 11;
    const DISABLE_SVC_FLAP_DETECTION = 12;
    const ENABLE_SVC_FLAP_DETECTION = 13;
    const DISABLE_SVC_EVENT_HANDLER = 14;
    const ENABLE_SVC_EVENT_HANDLER = 15;
    const DISABLE_SVC_CHECK = 16;
    const ENABLE_SVC_CHECK = 17;
    const DISABLE_PASSIVE_SVC_CHECKS = 18;
    const ENABLE_PASSIVE_SVC_CHECKS = 19;
    const DISABLE_HOST_CHECK = 20;
    const ENABLE_HOST_CHECK = 21;
    const DISABLE_HOST_EVENT_HANDLER = 22;
    const ENABLE_HOST_EVENT_HANDLER = 23;
    const DISABLE_HOST_FLAP_DETECTION = 24;
    const ENABLE_HOST_FLAP_DETECTION = 25;
    const DISABLE_HOST_NOTIFICATIONS = 26;
    const ENABLE_HOST_NOTIFICATIONS = 27;
    const DISABLE_PASSIVE_HOST_CHECKS = 28;
    const ENABLE_PASSIVE_HOST_CHECKS = 29;

    /**
     * @var ExternalCommandQueryOptions
     */
    private $ExternalCommandQueryOptions;

    /**
     * ExternalCommand constructor.
     * @param ExternalCommandQueryOptions $ExternalCommandQueryOptions
     */
    public function __construct(ExternalCommandQueryOptions $ExternalCommandQueryOptions) {
        $this->ExternalCommandQueryOptions = $ExternalCommandQueryOptions;
    }

    /**
     * @return string|array
     */
    public function getCommand() {
        switch ($this->ExternalCommandQueryOptions->getCommand()) {
            case self::ENABLE_HOST_NOTIFICATIONS:
                return sprintf('[%s] ENABLE_HOST_NOTIFICATIONS;%s', time(), $this->ExternalCommandQueryOptions->getHostname());
                break;

            case self::DISABLE_HOST_NOTIFICATIONS:
                return sprintf('[%s] DISABLE_HOST_NOTIFICATIONS;%s', time(), $this->ExternalCommandQueryOptions->getHostname());
                break;

            case self::ENABLE_HOST_CHECK:
                return sprintf('[%s] ENABLE_HOST_CHECK;%s', time(), $this->ExternalCommandQueryOptions->getHostname());
                break;

            case self::DISABLE_HOST_CHECK:
                return sprintf('[%s] DISABLE_HOST_CHECK;%s', time(), $this->ExternalCommandQueryOptions->getHostname());
                break;

            case self::ENABLE_PASSIVE_HOST_CHECKS:
                return sprintf('[%s] ENABLE_PASSIVE_HOST_CHECKS;%s', time(), $this->ExternalCommandQueryOptions->getHostname());
                break;

            case self::DISABLE_PASSIVE_HOST_CHECKS:
                return sprintf('[%s] DISABLE_PASSIVE_HOST_CHECKS;%s', time(), $this->ExternalCommandQueryOptions->getHostname());
                break;

            case self::ENABLE_HOST_FLAP_DETECTION:
                return sprintf('[%s] ENABLE_HOST_FLAP_DETECTION;%s', time(), $this->ExternalCommandQueryOptions->getHostname());
                break;

            case self::DISABLE_HOST_FLAP_DETECTION:
                return sprintf('[%s] DISABLE_HOST_FLAP_DETECTION;%s', time(), $this->ExternalCommandQueryOptions->getHostname());
                break;

            case self::ENABLE_HOST_EVENT_HANDLER:
                return sprintf('[%s] ENABLE_HOST_EVENT_HANDLER;%s', time(), $this->ExternalCommandQueryOptions->getHostname());
                break;

            case self::DISABLE_HOST_EVENT_HANDLER:
                return sprintf('[%s] DISABLE_HOST_EVENT_HANDLER;%s', time(), $this->ExternalCommandQueryOptions->getHostname());
                break;

            case self::SCHEDULE_FORCED_HOST_CHECK:
                return sprintf('[%s] SCHEDULE_FORCED_HOST_CHECK;%s;%s', time(), $this->ExternalCommandQueryOptions->getHostname(), time());
                break;

            case self::SCHEDULE_FORCED_HOST_SVC_CHECKS:
                $commands = [];
                $commands[] = sprintf('[%s] SCHEDULE_FORCED_HOST_CHECK;%s;%s', time(), $this->ExternalCommandQueryOptions->getHostname(), time());
                $commands[] = sprintf('[%s] SCHEDULE_FORCED_HOST_SVC_CHECKS;%s;%s', time(), $this->ExternalCommandQueryOptions->getHostname(), time());
                return $commands;
                break;

            case self::ENABLE_SVC_NOTIFICATIONS:
                return sprintf('[%s] ENABLE_SVC_NOTIFICATIONS;%s;%s',
                    time(),
                    $this->ExternalCommandQueryOptions->getHostname(),
                    $this->ExternalCommandQueryOptions->getServicedescription()
                );
                break;

            case self::DISABLE_SVC_NOTIFICATIONS:
                return sprintf('[%s] DISABLE_SVC_NOTIFICATIONS;%s;%s',
                    time(),
                    $this->ExternalCommandQueryOptions->getHostname(),
                    $this->ExternalCommandQueryOptions->getServicedescription()
                );
                break;

            case self::ENABLE_SVC_CHECK:
                return sprintf('[%s] ENABLE_SVC_CHECK;%s;%s',
                    time(),
                    $this->ExternalCommandQueryOptions->getHostname(),
                    $this->ExternalCommandQueryOptions->getServicedescription()
                );
                break;

            case self::DISABLE_SVC_CHECK:
                return sprintf('[%s] DISABLE_SVC_CHECK;%s;%s',
                    time(),
                    $this->ExternalCommandQueryOptions->getHostname(),
                    $this->ExternalCommandQueryOptions->getServicedescription()
                );
                break;

            case self::ENABLE_PASSIVE_SVC_CHECKS:
                return sprintf('[%s] ENABLE_PASSIVE_SVC_CHECKS;%s;%s',
                    time(),
                    $this->ExternalCommandQueryOptions->getHostname(),
                    $this->ExternalCommandQueryOptions->getServicedescription()
                );
                break;

            case self::DISABLE_PASSIVE_SVC_CHECKS:
                return sprintf('[%s] DISABLE_PASSIVE_SVC_CHECKS;%s;%s',
                    time(),
                    $this->ExternalCommandQueryOptions->getHostname(),
                    $this->ExternalCommandQueryOptions->getServicedescription()
                );
                break;

            case self::ENABLE_SVC_FLAP_DETECTION:
                return sprintf('[%s] ENABLE_SVC_FLAP_DETECTION;%s;%s',
                    time(),
                    $this->ExternalCommandQueryOptions->getHostname(),
                    $this->ExternalCommandQueryOptions->getServicedescription()
                );
                break;

            case self::DISABLE_SVC_FLAP_DETECTION:
                return sprintf('[%s] DISABLE_SVC_FLAP_DETECTION;%s;%s',
                    time(),
                    $this->ExternalCommandQueryOptions->getHostname(),
                    $this->ExternalCommandQueryOptions->getServicedescription()
                );
                break;

            case self::ENABLE_SVC_EVENT_HANDLER:
                return sprintf('[%s] ENABLE_SVC_EVENT_HANDLER;%s;%s',
                    time(),
                    $this->ExternalCommandQueryOptions->getHostname(),
                    $this->ExternalCommandQueryOptions->getServicedescription()
                );
                break;

            case self::DISABLE_SVC_EVENT_HANDLER:
                return sprintf('[%s] DISABLE_SVC_EVENT_HANDLER;%s;%s',
                    time(),
                    $this->ExternalCommandQueryOptions->getHostname(),
                    $this->ExternalCommandQueryOptions->getServicedescription()
                );
                break;


            case self::SCHEDULE_FORCED_SVC_CHECK:
                return sprintf('[%s] SCHEDULE_FORCED_SVC_CHECK;%s;%s;%s',
                    time(),
                    $this->ExternalCommandQueryOptions->getHostname(),
                    $this->ExternalCommandQueryOptions->getServicedescription(),
                    time()
                );
                break;

            default:
                return '';
        }
    }

}
