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

namespace Statusengine\Controller;

use Statusengine\Loader\Crate\ExternalCommandSaver;
use Statusengine\Loader\HostDowntimeLoaderInterface;
use Statusengine\Loader\ScheduleddowntimeHostLoaderInterface;
use Statusengine\Loader\ScheduleddowntimeServiceLoaderInterface;
use Statusengine\Loader\ServiceDowntimeLoaderInterface;
use Statusengine\Saver\ExternalCommandSaverInterface;
use Statusengine\ValueObjects\ExternalCommandArgsQueryOptions;
use Statusengine\ValueObjects\ExternalCommandQueryOptions;

class ExternalCommand extends Controller {

    /**
     * @var ExternalCommandSaver
     */
    private $ExternalCommandSaver;

    /**
     * ExternalCommand constructor.
     * @param ExternalCommandSaverInterface $ExternalCommandSaver
     */
    public function __construct(ExternalCommandSaverInterface $ExternalCommandSaver) {
        $this->ExternalCommandSaver = $ExternalCommandSaver;
    }

    /**
     * @param ExternalCommandQueryOptions $ExternalCommandQueryOptions
     * @return bool
     */
    public function index(ExternalCommandQueryOptions $ExternalCommandQueryOptions) {
        $ExternalCommand = new \Statusengine\Generators\ExternalCommand($ExternalCommandQueryOptions);
        $commandAsStringOrArray = $ExternalCommand->getCommand();
        if(is_array($commandAsStringOrArray)){
            foreach($commandAsStringOrArray as $command){
                $this->ExternalCommandSaver->saveCommand($command, $ExternalCommandQueryOptions->getNodeName());
            }
            return true;
        }

        return $this->ExternalCommandSaver->saveCommand($commandAsStringOrArray, $ExternalCommandQueryOptions->getNodeName());

    }

    /**
     * @param ExternalCommandArgsQueryOptions $ExternalCommandArgsQueryOptions
     * @return bool
     */
    public function args(ExternalCommandArgsQueryOptions $ExternalCommandArgsQueryOptions) {
        $ExternalCommand = new \Statusengine\Generators\ExternalCommandArgs($ExternalCommandArgsQueryOptions);
        $commandAsStringOrArray = $ExternalCommand->getCommand();
        if(is_array($commandAsStringOrArray)){
            foreach($commandAsStringOrArray as $command){
                $this->ExternalCommandSaver->saveCommand($command, $ExternalCommandArgsQueryOptions->getNodeName());
            }
            return true;
        }

        return $this->ExternalCommandSaver->saveCommand($commandAsStringOrArray, $ExternalCommandArgsQueryOptions->getNodeName());

    }

    public function deleteHostIncludingServiceDowntimes(ExternalCommandArgsQueryOptions $ExternalCommandArgsQueryOptions, ScheduleddowntimeHostLoaderInterface $ScheduleddowntimeHostLoader, ScheduleddowntimeServiceLoaderInterface $ScheduleddowntimeServiceLoader){
        $hostDowntime = $ScheduleddowntimeHostLoader->getScheduledHostdowntimeById($ExternalCommandArgsQueryOptions->getDowntimeId());
        if(!empty($hostDowntime)){
            $serviceDowntimes = $ScheduleddowntimeServiceLoader->getScheduledServicedowntimesByHostdowntime($hostDowntime[0]);

            //Delete host downtime
            $command = sprintf('[%s] DEL_HOST_DOWNTIME;%s', time(), $ExternalCommandArgsQueryOptions->getDowntimeId());
            $this->ExternalCommandSaver->saveCommand($command, $ExternalCommandArgsQueryOptions->getNodeName());

            //Delete service downtimes
            foreach($serviceDowntimes as $serviceDowntime){
                $command = sprintf('[%s] DEL_SVC_DOWNTIME;%s', time(), $serviceDowntime['internal_downtime_id']);
                $this->ExternalCommandSaver->saveCommand($command, $ExternalCommandArgsQueryOptions->getNodeName());
            }
        }
    }

}
