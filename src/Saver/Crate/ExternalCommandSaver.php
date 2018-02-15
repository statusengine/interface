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

namespace Statusengine\Loader\Crate;

use Crate\PDO\PDO;
use Statusengine\Backend\Crate\Crate;
use Statusengine\Backend\StorageBackend;
use Statusengine\Generators\Uuid;
use Statusengine\Saver\ExternalCommandSaverInterface;

class ExternalCommandSaver implements ExternalCommandSaverInterface {

    /**
     * @var StorageBackend
     */
    private $StorageBackend;

    /**
     * @var \Statusengine\Backend\Crate\Crate
     */
    private $Backend;

    /**
     * ExternalCommandSaver constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }

    /**
     * @param string $command
     * @param string $nodeName
     * @return bool
     */
    public function saveCommand($command, $nodeName){
        $baseQuery = 'INSERT INTO statusengine_tasks (entry_time, node_name, payload, type, uuid)VALUES(?,?,?,?,?)';
        $query = $this->Backend->prepare($baseQuery);
        $query->bindValue(1, time());
        $query->bindValue(2, $nodeName);
        $query->bindValue(3, $command);
        $query->bindValue(4, 'externalcommand');
        $query->bindValue(5, Uuid::v4());

        return $this->Backend->executeQuery($query);
    }

}
