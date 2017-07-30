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

namespace Statusengine\Loader\Crate;


use Statusengine\Backend\Crate\Crate;
use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\DashboardLoaderInterface;

class DashboardLoader implements DashboardLoaderInterface {

    /**
     * @var StorageBackend
     */
    private $StorageBackend;

    /**
     * @var \Statusengine\Backend\Crate\Crate
     */
    private $Backend;

    /**
     * DashboardLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }


    /**
     * @return int
     */
    public function getNumberOfMonitoredHosts() {
        $query = $this->Backend->prepare('SELECT COUNT(*) as count FROM statusengine_hoststatus');
        $result = $this->Backend->fetchAll($query);
        return (int)$result[0]['count'];
    }

    /**
     * @return int
     */
    public function getNumberOfMonitoredServices() {
        $query = $this->Backend->prepare('SELECT COUNT(*) as count FROM statusengine_servicestatus');
        $result = $this->Backend->fetchAll($query);
        return (int)$result[0]['count'];
    }

    /**
     * @return array
     */
    public function getHostOverview() {
        $result = [
            0 => 0, //up
            1 => 0, //down
            2 => 0, //unreachable
        ];
        $query = $this->Backend->prepare('SELECT current_state, COUNT(current_state) AS count FROM statusengine_hoststatus GROUP BY current_state');
        foreach ($this->Backend->fetchAll($query) as $currentState) {
            $result[$currentState['current_state']] = $currentState['count'];
        }

        return $result;
    }

    /**
     * @return array
     */
    public function getServiceOverview() {
        $result = [
            0 => 0, //ok
            1 => 0, //warning
            2 => 0, //critical
            3 => 0, //unknown
        ];
        $query = $this->Backend->prepare('SELECT current_state, COUNT(current_state) AS count FROM statusengine_servicestatus GROUP BY current_state');
        foreach ($this->Backend->fetchAll($query) as $currentState) {
            $result[$currentState['current_state']] = $currentState['count'];
        }

        return $result;
    }

    /**
     * @return int
     */
    public function getNumberOfServiceProblems() {
        $query = $this->Backend->prepare('
          SELECT COUNT(*) AS count FROM statusengine_servicestatus 
          WHERE current_state > 0
          AND problem_has_been_acknowledged = 0
          AND scheduled_downtime_depth = 0
        ');
        $result = $this->Backend->fetchAll($query);
        return (int)$result[0]['count'];
    }

}