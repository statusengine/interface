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

use Crate\PDO\PDO;
use Statusengine\Backend\Crate\Crate;
use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\ClusterLoaderInterface;

class ClusterLoader implements ClusterLoaderInterface {

    /**
     * @var StorageBackend
     */
    private $StorageBackend;

    /**
     * @var \Statusengine\Backend\Crate\Crate
     */
    private $Backend;

    /**
     * ClusterLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }


    /**
     * @return array
     */
    public function getClusterNodes() {
        return $this->baseQuery(['node_name']);
    }

    public function getClusterOverview() {
        $clusterNodes = $this->baseQuery(['*']);
        foreach ($clusterNodes as $key => $node) {
            $hostCount = $this->Backend->prepare('SELECT COUNT(*) as counter from statusengine_hoststatus where node_name=?');
            $serviceCount = $this->Backend->prepare('SELECT COUNT(*) as counter from statusengine_servicestatus where node_name=?');
            $hostCount->bindValue(1, $node['node_name']);
            $serviceCount->bindValue(1, $node['node_name']);
            $hostCount = $this->Backend->fetchAll($hostCount);
            $serviceCount = $this->Backend->fetchAll($serviceCount);

            $clusterNodes[$key]['number_of_hosts'] = $hostCount[0]['counter'];
            $clusterNodes[$key]['number_of_services'] = $serviceCount[0]['counter'];
        }

        return $clusterNodes;
    }

    /**
     * @param array $fields
     * @return array
     */
    private function baseQuery($fields = []) {
        $query = $this->Backend->prepare(
            sprintf('SELECT %s FROM statusengine_nodes ORDER BY node_name ASC', implode(',', $fields))
        );
        return $this->Backend->fetchAll($query);
    }

}
